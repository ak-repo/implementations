// Blog Search Frontend Application
const API_BASE_URL = ''; // Same origin

// State
let currentPage = 1;
let currentQuery = '';
let autocompleteTimeout = null;
let currentResults = [];

// DOM Elements - Search
const searchInput = document.getElementById('searchInput');
const searchBtn = document.getElementById('searchBtn');
const authorFilter = document.getElementById('authorFilter');
const tagFilter = document.getElementById('tagFilter');
const sortBy = document.getElementById('sortBy');
const autocompleteDropdown = document.getElementById('autocompleteDropdown');

const resultsSection = document.getElementById('resultsSection');
const resultsList = document.getElementById('resultsList');
const resultsStats = document.getElementById('resultsStats');
const pageInfo = document.getElementById('pageInfo');
const pagination = document.getElementById('pagination');

const emptyState = document.getElementById('emptyState');
const loadingState = document.getElementById('loadingState');
const noResults = document.getElementById('noResults');

// DOM Elements - Create Blog
const createBlogBtn = document.getElementById('createBlogBtn');
const createBlogModal = document.getElementById('createBlogModal');
const closeModalBtn = document.getElementById('closeModalBtn');
const cancelBtn = document.getElementById('cancelBtn');
const createBlogForm = document.getElementById('createBlogForm');
const blogTitle = document.getElementById('blogTitle');
const blogContent = document.getElementById('blogContent');
const blogAuthor = document.getElementById('blogAuthor');
const blogTags = document.getElementById('blogTags');
const titleCharCount = document.getElementById('titleCharCount');
const formError = document.getElementById('formError');
const errorText = document.getElementById('errorText');
const submitBtn = document.getElementById('submitBtn');
const submitText = document.getElementById('submitText');
const submitSpinner = document.getElementById('submitSpinner');
const successToast = document.getElementById('successToast');
const toastMessage = document.getElementById('toastMessage');

// Event Listeners
document.addEventListener('DOMContentLoaded', () => {
    searchBtn.addEventListener('click', handleSearch);
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            handleSearch();
        }
    });
    
    // Autocomplete with debounce
    searchInput.addEventListener('input', (e) => {
        const query = e.target.value.trim();
        
        // Clear existing timeout
        if (autocompleteTimeout) {
            clearTimeout(autocompleteTimeout);
        }
        
        // Hide dropdown if query is empty
        if (query.length < 2) {
            hideAutocomplete();
            return;
        }
        
        // Debounce autocomplete
        autocompleteTimeout = setTimeout(() => {
            fetchAutocomplete(query);
        }, 200);
    });
    
    // Hide autocomplete on click outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.relative')) {
            hideAutocomplete();
        }
    });
    
    // Filter changes trigger search
    authorFilter.addEventListener('change', handleSearch);
    tagFilter.addEventListener('change', handleSearch);
    sortBy.addEventListener('change', handleSearch);
    
    // Create Blog Modal Events
    createBlogBtn.addEventListener('click', openCreateModal);
    closeModalBtn.addEventListener('click', closeCreateModal);
    cancelBtn.addEventListener('click', closeCreateModal);
    createBlogModal.addEventListener('click', (e) => {
        if (e.target === createBlogModal) closeCreateModal();
    });
    
    // Form submission
    createBlogForm.addEventListener('submit', handleCreateBlog);
    
    // Title character count
    blogTitle.addEventListener('input', () => {
        titleCharCount.textContent = blogTitle.value.length;
    });
    
    // Enter key in modal
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !createBlogModal.classList.contains('hidden')) {
            closeCreateModal();
        }
    });
});

// Handle Search
async function handleSearch() {
    currentQuery = searchInput.value.trim();
    currentPage = 1;
    await performSearch();
}

// Perform Search API Call
async function performSearch() {
    const author = authorFilter.value.trim();
    const tag = tagFilter.value.trim();
    const sort = sortBy.value;
    
    // Show loading
    showLoading();
    hideAutocomplete();
    
    try {
        const params = new URLSearchParams();
        if (currentQuery) params.append('q', currentQuery);
        if (author) params.append('author', author);
        if (tag) params.append('tag', tag);
        params.append('page', currentPage.toString());
        params.append('size', '10');
        params.append('sort', sort);
        
        const response = await fetch(`${API_BASE_URL}/search?${params}`);
        
        if (!response.ok) {
            throw new Error('Search failed');
        }
        
        const data = await response.json();
        currentResults = data.blogs || [];
        
        renderResults(data);
    } catch (error) {
        console.error('Search error:', error);
        showError('Failed to perform search. Please try again.');
    }
}

// Fetch Autocomplete Suggestions
async function fetchAutocomplete(query) {
    if (query.length < 2) return;
    
    try {
        const response = await fetch(`${API_BASE_URL}/autocomplete?q=${encodeURIComponent(query)}`);
        
        if (!response.ok) {
            throw new Error('Autocomplete failed');
        }
        
        const data = await response.json();
        renderAutocomplete(data.suggestions || []);
    } catch (error) {
        console.error('Autocomplete error:', error);
    }
}

// Render Autocomplete Dropdown
function renderAutocomplete(suggestions) {
    if (suggestions.length === 0) {
        hideAutocomplete();
        return;
    }
    
    const html = suggestions.map(suggestion => `
        <div 
            class="autocomplete-item px-4 py-2 cursor-pointer text-gray-700"
            onclick="selectSuggestion('${escapeHtml(suggestion)}')"
        >
            ${escapeHtml(suggestion)}
        </div>
    `).join('');
    
    autocompleteDropdown.innerHTML = html;
    autocompleteDropdown.classList.remove('hidden');
}

// Hide Autocomplete
function hideAutocomplete() {
    autocompleteDropdown.classList.add('hidden');
}

// Select Autocomplete Suggestion
function selectSuggestion(suggestion) {
    searchInput.value = suggestion;
    hideAutocomplete();
    handleSearch();
}

// Render Search Results
function renderResults(data) {
    hideLoading();
    
    const total = data.total || 0;
    const page = data.page || 1;
    const size = data.size || 10;
    const blogs = data.blogs || [];
    
    // Update stats
    resultsStats.textContent = `${total.toLocaleString()} result${total !== 1 ? 's' : ''} found`;
    
    const totalPages = Math.ceil(total / size);
    pageInfo.textContent = `Page ${page} of ${totalPages}`;
    
    // Show/hide sections
    emptyState.classList.add('hidden');
    
    if (blogs.length === 0) {
        resultsSection.classList.add('hidden');
        noResults.classList.remove('hidden');
        return;
    }
    
    noResults.classList.add('hidden');
    resultsSection.classList.remove('hidden');
    
    // Render blog cards
    resultsList.innerHTML = blogs.map(blog => `
        <article class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
            <div class="flex justify-between items-start mb-2">
                <h2 class="text-xl font-semibold text-gray-900 leading-tight">
                    ${blog.title}
                </h2>
            </div>
            
            <div class="flex items-center gap-4 text-sm text-gray-500 mb-3">
                <span class="font-medium text-gray-700">${escapeHtml(blog.author)}</span>
                <span>•</span>
                <span>${formatDate(blog.created_at)}</span>
            </div>
            
            <p class="text-gray-600 leading-relaxed mb-4">
                ${truncateText(stripHtml(blog.content), 250)}
            </p>
            
            <div class="flex flex-wrap gap-2">
                ${blog.tags.map(tag => `
                    <span class="px-2 py-1 bg-blue-100 text-blue-700 text-xs font-medium rounded">
                        ${escapeHtml(tag)}
                    </span>
                `).join('')}
            </div>
        </article>
    `).join('');
    
    // Render pagination
    renderPagination(page, totalPages, total, size);
}

// Render Pagination
function renderPagination(currentPage, totalPages, total, pageSize) {
    if (totalPages <= 1) {
        pagination.innerHTML = '';
        return;
    }
    
    let html = '';
    
    // Previous button
    html += `
        <button 
            onclick="goToPage(${currentPage - 1})"
            class="px-3 py-2 rounded border ${currentPage === 1 ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'}"
            ${currentPage === 1 ? 'disabled' : ''}
        >
            ← Prev
        </button>
    `;
    
    // Page numbers
    const maxVisible = 5;
    let startPage = Math.max(1, currentPage - Math.floor(maxVisible / 2));
    let endPage = Math.min(totalPages, startPage + maxVisible - 1);
    
    if (endPage - startPage + 1 < maxVisible) {
        startPage = Math.max(1, endPage - maxVisible + 1);
    }
    
    if (startPage > 1) {
        html += `<button onclick="goToPage(1)" class="px-3 py-2 rounded border bg-white text-gray-700 hover:bg-gray-50">1</button>`;
        if (startPage > 2) {
            html += `<span class="px-2 text-gray-500">...</span>`;
        }
    }
    
    for (let i = startPage; i <= endPage; i++) {
        html += `
            <button 
                onclick="goToPage(${i})"
                class="px-3 py-2 rounded border ${i === currentPage ? 'bg-blue-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50'}"
            >
                ${i}
            </button>
        `;
    }
    
    if (endPage < totalPages) {
        if (endPage < totalPages - 1) {
            html += `<span class="px-2 text-gray-500">...</span>`;
        }
        html += `<button onclick="goToPage(${totalPages})" class="px-3 py-2 rounded border bg-white text-gray-700 hover:bg-gray-50">${totalPages}</button>`;
    }
    
    // Next button
    html += `
        <button 
            onclick="goToPage(${currentPage + 1})"
            class="px-3 py-2 rounded border ${currentPage === totalPages ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'}"
            ${currentPage === totalPages ? 'disabled' : ''}
        >
            Next →
        </button>
    `;
    
    pagination.innerHTML = html;
}

// Go to specific page
function goToPage(page) {
    if (page < 1) return;
    currentPage = page;
    performSearch();
    // Scroll to top of results
    resultsSection.scrollIntoView({ behavior: 'smooth' });
}

// UI State Helpers
function showLoading() {
    emptyState.classList.add('hidden');
    resultsSection.classList.add('hidden');
    noResults.classList.add('hidden');
    loadingState.classList.remove('hidden');
}

function hideLoading() {
    loadingState.classList.add('hidden');
}

function showError(message) {
    hideLoading();
    resultsSection.classList.add('hidden');
    noResults.classList.remove('hidden');
    noResults.innerHTML = `
        <p class="text-red-500 text-lg">${escapeHtml(message)}</p>
        <p class="text-gray-400 mt-2">Please try again later.</p>
    `;
}

// Utility Functions
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function stripHtml(html) {
    const tmp = document.createElement('div');
    tmp.innerHTML = html;
    return tmp.textContent || tmp.innerText || '';
}

function truncateText(text, maxLength) {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
}

function formatDate(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
}

// ============================
// Create Blog Functions
// ============================

function openCreateModal() {
    createBlogModal.classList.remove('hidden');
    createBlogModal.classList.add('flex');
    document.body.style.overflow = 'hidden';
    
    // Focus on title
    setTimeout(() => blogTitle.focus(), 100);
}

function closeCreateModal() {
    createBlogModal.classList.add('hidden');
    createBlogModal.classList.remove('flex');
    document.body.style.overflow = '';
    
    // Reset form
    resetForm();
}

function resetForm() {
    createBlogForm.reset();
    titleCharCount.textContent = '0';
    hideFormError();
}

function showFormError(message) {
    formError.classList.remove('hidden');
    errorText.textContent = message;
}

function hideFormError() {
    formError.classList.add('hidden');
    errorText.textContent = '';
}

function setSubmitting(isSubmitting) {
    submitBtn.disabled = isSubmitting;
    if (isSubmitting) {
        submitText.textContent = 'Creating...';
        submitSpinner.classList.remove('hidden');
    } else {
        submitText.textContent = 'Create Blog';
        submitSpinner.classList.add('hidden');
    }
}

async function handleCreateBlog(e) {
    e.preventDefault();
    
    // Get form data
    const title = blogTitle.value.trim();
    const content = blogContent.value.trim();
    const author = blogAuthor.value.trim();
    const tagsInput = blogTags.value.trim();
    
    // Validation
    if (!title) {
        showFormError('Title is required');
        blogTitle.focus();
        return;
    }
    if (!content) {
        showFormError('Content is required');
        blogContent.focus();
        return;
    }
    if (!author) {
        showFormError('Author is required');
        blogAuthor.focus();
        return;
    }
    
    // Parse tags
    const tags = tagsInput
        ? tagsInput.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0)
        : [];
    
    hideFormError();
    setSubmitting(true);
    
    try {
        const response = await fetch(`${API_BASE_URL}/blogs`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                title,
                content,
                author,
                tags
            })
        });
        
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || errorData.message || 'Failed to create blog');
        }
        
        const data = await response.json();
        
        // Success
        closeCreateModal();
        showToast('Blog created successfully!');
        
        // Refresh search to show new blog
        currentQuery = title;
        searchInput.value = title;
        currentPage = 1;
        await performSearch();
        
    } catch (error) {
        console.error('Create blog error:', error);
        showFormError(error.message || 'Failed to create blog. Please try again.');
    } finally {
        setSubmitting(false);
    }
}

function showToast(message) {
    toastMessage.textContent = message;
    successToast.classList.remove('translate-y-20', 'opacity-0');
    
    setTimeout(() => {
        successToast.classList.add('translate-y-20', 'opacity-0');
    }, 3000);
}