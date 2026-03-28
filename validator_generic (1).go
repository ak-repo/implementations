// Generic CSV Upload Validator
// ==============================
// Validates any data CSV against a template CSV.
// Columns in the template = expected upload columns.
// Required by default — unless listed in --skip (comma-separated).
//
// Usage:
//   go run validator_generic.go \
//     --template ProductTemplate.csv \
//     --data    product__6_.csv \
//     --skip    name_ar,multipacks,allow_decimal_quantity,decimal_calculation
//
// Optional flags:
//   --out-safe      output file for valid rows       (default: <data>_safe.csv)
//   --out-rejected  output file for rejected rows    (default: <data>_rejected.csv)
//   --master        master/reference CSV file        (optional)
//   --master-col    column name inside master CSV to validate against (required if --master set)
//   --validate-col  column in data CSV to check against master        (required if --master set)
//
// Master lookup example (validate category_code against a category master):
//   --master productCategory.csv --master-col code --validate-col category_code
//
// Multiple master lookups: repeat the three flags as pairs — not yet supported
// in this version (one master lookup per run). Chain runs for multiple lookups.
//
// Output files:
//   *_safe.csv     — template columns ONLY, ready to upload
//   *_rejected.csv — _row_number + _reject_reason prepended, then template columns

package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// utf8BOM is the byte order mark that Excel silently adds to CSV files.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// newCSVReader wraps a file in a csv.Reader, stripping a UTF-8 BOM if present.
// This fixes the silent failure where Excel-exported CSVs cause the first
// column header to read as "\xef\xbb\xbfname" instead of "name", making
// every row appear to be missing required fields.
func newCSVReader(f *os.File) *csv.Reader {
	bom := make([]byte, 3)
	n, _ := f.Read(bom)
	if n < 3 || !bytes.Equal(bom[:3], utf8BOM) {
		f.Seek(0, io.SeekStart) // no BOM — rewind
	}
	// if BOM found, file position is already past it — just continue
	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	return r
}

// ─────────────────────────────────────────────────────────────
// CSV HELPERS
// ─────────────────────────────────────────────────────────────

// readCSVHeaders reads only the header row from a CSV file.
func readCSVHeaders(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %q: %w", path, err)
	}
	defer f.Close()

	r := newCSVReader(f)

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("cannot read header from %q: %w", path, err)
	}
	trimmed := make([]string, len(header))
	for i, h := range header {
		trimmed[i] = strings.TrimSpace(h)
	}
	return trimmed, nil
}

// streamCSV opens a CSV and returns the header + a row iterator.
// The iterator returns (rowMap, rowIndex, error). rowIndex is 1-based
// counting from the first data row (so row 2 in the file).
// Returns io.EOF when done. Caller must close the returned *os.File.
func streamCSV(path string) ([]string, func() (map[string]string, int, error), *os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot open %q: %w", path, err)
	}

	r := newCSVReader(f)

	header, err := r.Read()
	if err != nil {
		f.Close()
		return nil, nil, nil, fmt.Errorf("cannot read header from %q: %w", path, err)
	}
	trimmedHeader := make([]string, len(header))
	for i, h := range header {
		trimmedHeader[i] = strings.TrimSpace(h)
	}

	lineNum := 1 // header was line 1
	iter := func() (map[string]string, int, error) {
		rec, err := r.Read()
		if err != nil {
			return nil, 0, err // io.EOF when done
		}
		lineNum++
		row := make(map[string]string, len(trimmedHeader))
		for i, h := range trimmedHeader {
			val := ""
			if i < len(rec) {
				val = strings.TrimSpace(rec[i])
			}
			row[h] = val
		}
		return row, lineNum, nil
	}

	return trimmedHeader, iter, f, nil
}

// buildLookupSet reads a CSV and collects all unique values from one column.
func buildLookupSet(path, col string) (map[string]struct{}, error) {
	_, iter, f, err := streamCSV(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	set := make(map[string]struct{})
	for {
		row, _, err := iter()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading master %q: %w", path, err)
		}
		val := strings.TrimSpace(row[col])
		if val != "" {
			set[val] = struct{}{}
		}
	}
	return set, nil
}

// writeCSVFromMaps writes rows (as maps) to a CSV with exact column order.
func writeCSVFromMaps(path string, cols []string, rows []map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create %q: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(cols); err != nil {
		return err
	}
	rec := make([]string, len(cols))
	for _, row := range rows {
		for i, col := range cols {
			rec[i] = row[col]
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	return w.Error()
}

// ─────────────────────────────────────────────────────────────
// VALIDATION CORE
// ─────────────────────────────────────────────────────────────

// isEmpty returns true for blank strings or the literal "nan".
func isEmpty(s string) bool {
	l := strings.ToLower(strings.TrimSpace(s))
	return l == "" || l == "nan"
}

// ValidationConfig holds everything needed for one validation run.
type ValidationConfig struct {
	TemplatePath string            // defines column order
	DataPath     string            // rows to validate
	SkipCols     map[string]bool   // columns to skip validation on
	MasterLookup *MasterLookup     // optional reference check
	OutSafe      string            // output path for safe rows
	OutRejected  string            // output path for rejected rows
}

// MasterLookup describes one reference-integrity check.
type MasterLookup struct {
	MasterPath string // path to the master/reference CSV
	MasterCol  string // column in master CSV holding valid values
	DataCol    string // column in data CSV to validate
}

// ValidationResult holds counters and top rejection reasons.
type ValidationResult struct {
	TotalRows     int
	AcceptedCount int
	RejectedCount int
	TopReasons    []ReasonCount
}

// ReasonCount is a reason + how many rows had it.
type ReasonCount struct {
	Reason string
	Count  int
}

// RunValidation is the generic validation function.
// It reads the template to get column order, streams the data CSV,
// validates each row, and writes two output files.
func RunValidation(cfg ValidationConfig) (ValidationResult, error) {
	// ── 1. Read template headers (column order + required cols) ──
	templateCols, err := readCSVHeaders(cfg.TemplatePath)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("template error: %w", err)
	}
	if len(templateCols) == 0 {
		return ValidationResult{}, fmt.Errorf("template %q has no columns", cfg.TemplatePath)
	}

	// Build set of required columns (all template cols minus skipped ones)
	requiredCols := make(map[string]bool, len(templateCols))
	for _, col := range templateCols {
		if !cfg.SkipCols[col] {
			requiredCols[col] = true
		}
	}

	// ── 2. Load master lookup set (if provided) ───────────────────
	var lookupSet map[string]struct{}
	if cfg.MasterLookup != nil {
		lookupSet, err = buildLookupSet(cfg.MasterLookup.MasterPath, cfg.MasterLookup.MasterCol)
		if err != nil {
			return ValidationResult{}, fmt.Errorf("master lookup error: %w", err)
		}
	}

	// ── 3. Stream data CSV and validate ───────────────────────────
	_, iter, dataFile, err := streamCSV(cfg.DataPath)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("data file error: %w", err)
	}
	defer dataFile.Close()

	var acceptedRows []map[string]string
	var rejectedRows []map[string]string
	reasonCounts := make(map[string]int)
	total := 0

	for {
		row, lineNum, err := iter()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ValidationResult{}, fmt.Errorf("error reading data at line %d: %w", lineNum, err)
		}
		total++

		var errors []string

		// ── Null checks on required columns ──────────────────────
		for _, col := range templateCols {
			if !requiredCols[col] {
				continue // skip — user said this column is optional
			}
			if isEmpty(row[col]) {
				errors = append(errors, fmt.Sprintf("Missing required field: %q", col))
			}
		}

		// ── Master lookup check ───────────────────────────────────
		if cfg.MasterLookup != nil && lookupSet != nil {
			val := strings.TrimSpace(row[cfg.MasterLookup.DataCol])
			if !isEmpty(val) {
				if _, found := lookupSet[val]; !found {
					errors = append(errors,
						fmt.Sprintf("Invalid %q = %q — not found in master %q",
							cfg.MasterLookup.DataCol, val,
							filepath.Base(cfg.MasterLookup.MasterPath)))
				}
			}
		}

		// ── Build output row (template cols only) ─────────────────
		out := make(map[string]string, len(templateCols)+2)
		for _, col := range templateCols {
			v := row[col]
			if isEmpty(v) {
				v = ""
			}
			out[col] = v
		}

		if len(errors) > 0 {
			// Rejected: prepend reason columns
			out["_row_number"]    = fmt.Sprintf("%d", lineNum)
			out["_reject_reason"] = strings.Join(errors, " | ")
			rejectedRows = append(rejectedRows, out)
			for _, e := range errors {
				reasonCounts[e]++
			}
		} else {
			// Safe: template columns only — nothing extra
			acceptedRows = append(acceptedRows, out)
		}
	}

	// ── 4. Write output files ─────────────────────────────────────
	if err := writeCSVFromMaps(cfg.OutSafe, templateCols, acceptedRows); err != nil {
		return ValidationResult{}, fmt.Errorf("error writing safe file: %w", err)
	}

	rejCols := append([]string{"_row_number", "_reject_reason"}, templateCols...)
	if err := writeCSVFromMaps(cfg.OutRejected, rejCols, rejectedRows); err != nil {
		return ValidationResult{}, fmt.Errorf("error writing rejected file: %w", err)
	}

	// ── 5. Build result ───────────────────────────────────────────
	type kv struct{ k string; v int }
	var sorted []kv
	for k, v := range reasonCounts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })

	limit := 10
	if len(sorted) < limit {
		limit = len(sorted)
	}
	topReasons := make([]ReasonCount, limit)
	for i, item := range sorted[:limit] {
		topReasons[i] = ReasonCount{Reason: item.k, Count: item.v}
	}

	return ValidationResult{
		TotalRows:     total,
		AcceptedCount: len(acceptedRows),
		RejectedCount: len(rejectedRows),
		TopReasons:    topReasons,
	}, nil
}

// ─────────────────────────────────────────────────────────────
// OUTPUT HELPER — derives default output filenames
// ─────────────────────────────────────────────────────────────

func deriveOutputPaths(dataPath, outSafe, outRejected string) (string, string) {
	base := strings.TrimSuffix(dataPath, filepath.Ext(dataPath))
	if outSafe == "" {
		outSafe = base + "_safe.csv"
	}
	if outRejected == "" {
		outRejected = base + "_rejected.csv"
	}
	return outSafe, outRejected
}

// ─────────────────────────────────────────────────────────────
// CLI ENTRY POINT
// ─────────────────────────────────────────────────────────────

func main() {
	// ── Flags ─────────────────────────────────────────────────────
	templateFlag    := flag.String("template",     "", "Path to template CSV (defines column order & names)  [REQUIRED]")
	dataFlag        := flag.String("data",         "", "Path to data CSV to validate                         [REQUIRED]")
	skipFlag        := flag.String("skip",         "", "Comma-separated column names to skip validation on   [OPTIONAL]")
	outSafeFlag     := flag.String("out-safe",     "", "Output path for safe rows     (default: <data>_safe.csv)")
	outRejFlag      := flag.String("out-rejected",  "", "Output path for rejected rows (default: <data>_rejected.csv)")
	masterFlag      := flag.String("master",       "", "Path to master/reference CSV for lookup validation   [OPTIONAL]")
	masterColFlag   := flag.String("master-col",   "", "Column in master CSV holding valid values            [required with --master]")
	validateColFlag := flag.String("validate-col", "", "Column in data CSV to check against master           [required with --master]")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Generic CSV Upload Validator
=============================
Validates any data CSV against a template CSV.
All template columns are REQUIRED unless listed in --skip.

USAGE:
  go run validator_generic.go [flags]

FLAGS:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
EXAMPLES:

  # Basic — validate product CSV, skip optional fields
  go run validator_generic.go \
    --template ProductTemplate.csv \
    --data     product__6_.csv \
    --skip     name_ar,multipacks,allow_decimal_quantity,decimal_calculation

  # With master lookup — also check category_code against category master
  go run validator_generic.go \
    --template     ProductTemplate.csv \
    --data         product__6_.csv \
    --skip         name_ar,multipacks,allow_decimal_quantity,decimal_calculation \
    --master       productCategory.csv \
    --master-col   code \
    --validate-col category_code

  # Subcategory validation
  go run validator_generic.go \
    --template ProductSubcategoryTemplate.csv \
    --data     productSubCategory__1_.csv \
    --skip     name_ar,category_code

OUTPUT:
  <data>_safe.csv     — template columns only, upload ready
  <data>_rejected.csv — _row_number + _reject_reason + template columns
`)
	}

	flag.Parse()

	// ── Validate required flags ────────────────────────────────────
	if *templateFlag == "" || *dataFlag == "" {
		fmt.Fprintln(os.Stderr, "ERROR: --template and --data are required.")
		flag.Usage()
		os.Exit(1)
	}

	hasMaster := *masterFlag != ""
	if hasMaster && (*masterColFlag == "" || *validateColFlag == "") {
		fmt.Fprintln(os.Stderr, "ERROR: --master requires both --master-col and --validate-col.")
		os.Exit(1)
	}

	// ── Build skip set ─────────────────────────────────────────────
	skipCols := make(map[string]bool)
	if *skipFlag != "" {
		for _, col := range strings.Split(*skipFlag, ",") {
			col = strings.TrimSpace(col)
			if col != "" {
				skipCols[col] = true
			}
		}
	}

	// ── Derive output paths ────────────────────────────────────────
	outSafe, outRejected := deriveOutputPaths(*dataFlag, *outSafeFlag, *outRejFlag)

	// ── Build master lookup config ─────────────────────────────────
	var masterLookup *MasterLookup
	if hasMaster {
		masterLookup = &MasterLookup{
			MasterPath: *masterFlag,
			MasterCol:  *masterColFlag,
			DataCol:    *validateColFlag,
		}
	}

	// ── Print run header ───────────────────────────────────────────
	fmt.Println(strings.Repeat("=", 62))
	fmt.Printf("  GENERIC CSV VALIDATOR  —  %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 62))
	fmt.Printf("  Template     : %s\n", *templateFlag)
	fmt.Printf("  Data         : %s\n", *dataFlag)
	if len(skipCols) > 0 {
		skips := make([]string, 0, len(skipCols))
		for k := range skipCols {
			skips = append(skips, k)
		}
		sort.Strings(skips)
		fmt.Printf("  Skip cols    : %s\n", strings.Join(skips, ", "))
	}
	if masterLookup != nil {
		fmt.Printf("  Master check : %s [col=%s] → data col=%s\n",
			masterLookup.MasterPath, masterLookup.MasterCol, masterLookup.DataCol)
	}
	fmt.Printf("  Out safe     : %s\n", outSafe)
	fmt.Printf("  Out rejected : %s\n", outRejected)
	fmt.Println(strings.Repeat("─", 62))

	// ── Run validation ─────────────────────────────────────────────
	cfg := ValidationConfig{
		TemplatePath: *templateFlag,
		DataPath:     *dataFlag,
		SkipCols:     skipCols,
		MasterLookup: masterLookup,
		OutSafe:      outSafe,
		OutRejected:  outRejected,
	}

	result, err := RunValidation(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}

	// ── Print results ──────────────────────────────────────────────
	fmt.Printf("  Total rows   : %d\n", result.TotalRows)
	fmt.Printf("  ✅  Safe     : %d\n", result.AcceptedCount)
	fmt.Printf("  ❌  Rejected : %d\n", result.RejectedCount)

	if len(result.TopReasons) > 0 {
		fmt.Println("\n  Top rejection reasons:")
		for _, r := range result.TopReasons {
			fmt.Printf("    [%5dx]  %s\n", r.Count, r.Reason)
		}
	}

	fmt.Println(strings.Repeat("=", 62))
	fmt.Printf("  Done! Written to:\n")
	fmt.Printf("    %s\n", outSafe)
	fmt.Printf("    %s\n", outRejected)
	fmt.Println(strings.Repeat("=", 62))
}
