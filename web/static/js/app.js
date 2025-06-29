// App configuration
const CONFIG = {
  DEBOUNCE_DELAY: 300,
  MIN_SEARCH_LENGTH: 2,
  SEARCH_RESULTS_LIMIT: 5
};

// Global state
let searchTimeout = null;
let currentQuery = '';
let isSearching = false;

// DOM ready
document.addEventListener('DOMContentLoaded', function() {
  initializeApp();
});

function initializeApp() {
  // Initialize search functionality
  initializeSearch();
  
  // Initialize page-specific functionality
  const page = getPageType();
  switch(page) {
    case 'home':
      initializeHomePage();
      break;
    case 'jobs':
      initializeJobsPage();
      break;
    case 'job-detail':
      initializeJobDetailPage();
      break;
  }
}

function getPageType() {
  const path = window.location.pathname;
  if (path === '/') return 'home';
  if (path === '/jobs') return 'jobs';
  if (path.startsWith('/jobs/')) return 'job-detail';
  return 'unknown';
}

// Search functionality
function initializeSearch() {
  const searchInput = document.getElementById('search-input');
  const searchResults = document.getElementById('search-results');
  
  if (!searchInput) return;

  searchInput.addEventListener('input', handleSearchInput);
  searchInput.addEventListener('keydown', handleSearchKeydown);
  
  // Close search results when clicking outside
  document.addEventListener('click', function(e) {
    if (!searchInput.contains(e.target) && !searchResults.contains(e.target)) {
      hideSearchResults();
    }
  });
}

function handleSearchInput(e) {
  const query = e.target.value.trim();
  
  // Clear previous timeout
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  
  // If query is too short, hide results
  if (query.length < CONFIG.MIN_SEARCH_LENGTH) {
    hideSearchResults();
    currentQuery = '';
    return;
  }
  
  // Debounce the search
  searchTimeout = setTimeout(() => {
    performSearch(query);
  }, CONFIG.DEBOUNCE_DELAY);
}

function handleSearchKeydown(e) {
  if (e.key === 'Enter') {
    e.preventDefault();
    const query = e.target.value.trim();
    if (query) {
      navigateToJobsPage(query);
    }
  }
  
  if (e.key === 'Escape') {
    hideSearchResults();
  }
}

async function performSearch(query) {
  if (query === currentQuery && isSearching) return;
  
  currentQuery = query;
  isSearching = true;
  
  try {
    const response = await fetch(`/api/jobs/search?q=${encodeURIComponent(query)}&page_size=${CONFIG.SEARCH_RESULTS_LIMIT}`);
    const data = await response.json();
    
    if (data.success) {
      displaySearchResults(data.data.jobs, query);
    } else {
      showSearchError('Search failed. Please try again.');
    }
  } catch (error) {
    console.error('Search error:', error);
    showSearchError('Search failed. Please try again.');
  } finally {
    isSearching = false;
  }
}

function displaySearchResults(jobs, query) {
  const searchResults = document.getElementById('search-results');
  if (!searchResults) return;
  
  if (!jobs || jobs.length === 0) {
    searchResults.innerHTML = `
      <div class="search-result-item" style="text-align: center; color: var(--text-secondary);">
        No jobs found for "${query}"
      </div>
    `;
    searchResults.classList.add('show');
    return;
  }
  
  const resultsHTML = jobs.map(job => `
    <div class="search-result-item" onclick="navigateToJob(${job.id})">
      <div class="search-result-title">${escapeHtml(job.title)}</div>
      <div class="search-result-meta">${escapeHtml(job.companyName)} | ${escapeHtml(job.location)}</div>
    </div>
  `).join('');
  
  const seeAllButton = `
    <button class="see-all-btn" onclick="navigateToJobsPage('${escapeHtml(query)}')">
      See all results for "${query}"
    </button>
  `;
  
  searchResults.innerHTML = resultsHTML + seeAllButton;
  searchResults.classList.add('show');
}

function showSearchError(message) {
  const searchResults = document.getElementById('search-results');
  if (!searchResults) return;
  
  searchResults.innerHTML = `
    <div class="search-result-item error" style="text-align: center;">
      ${escapeHtml(message)}
    </div>
  `;
  searchResults.classList.add('show');
}

function hideSearchResults() {
  const searchResults = document.getElementById('search-results');
  if (searchResults) {
    searchResults.classList.remove('show');
  }
}

// Navigation functions
function navigateToJob(jobId) {
  window.location.href = `/jobs/${jobId}`;
}

function navigateToJobsPage(query = '') {
  const url = query ? `/jobs?q=${encodeURIComponent(query)}` : '/jobs';
  window.location.href = url;
}

// Home page functionality
function initializeHomePage() {
  // Focus on search input
  const searchInput = document.getElementById('search-input');
  if (searchInput) {
    searchInput.focus();
  }
}

// Jobs page functionality
function initializeJobsPage() {
  loadJobsFromURL();
  initializePagination();
}

async function loadJobsFromURL() {
  const urlParams = new URLSearchParams(window.location.search);
  const query = urlParams.get('q') || '';
  const page = parseInt(urlParams.get('page')) || 1;
  const pageSize = parseInt(urlParams.get('page_size')) || 20;
  
  // Set search input value if there's a query
  const searchInput = document.getElementById('search-input');
  if (searchInput && query) {
    searchInput.value = query;
  }
  
  try {
    await loadJobs(query, page, pageSize);
  } catch (error) {
    console.error('Error loading jobs:', error);
    showJobsError('Failed to load jobs. Please try again.');
  }
}

async function loadJobs(query = '', page = 1, pageSize = 20) {
  showJobsLoading();
  
  try {
    let url = `/api/jobs?page=${page}&page_size=${pageSize}`;
    if (query) {
      url = `/api/jobs/search?q=${encodeURIComponent(query)}&page=${page}&page_size=${pageSize}`;
    }
    
    const response = await fetch(url);
    const data = await response.json();
    
    if (data.success) {
      displayJobs(data.data.jobs);
      updateJobsHeader(data.data, query);
      updatePagination(data.data, query);
    } else {
      showJobsError(data.error || 'Failed to load jobs');
    }
  } catch (error) {
    console.error('Error loading jobs:', error);
    showJobsError('Failed to load jobs. Please try again.');
  }
}

function displayJobs(jobs) {
  const jobsList = document.getElementById('jobs-list');
  if (!jobsList) return;
  
  if (!jobs || jobs.length === 0) {
    jobsList.innerHTML = `
      <div class="error">
        <h3>No jobs found</h3>
        <p>Try adjusting your search criteria or check back later.</p>
      </div>
    `;
    return;
  }
  
  const jobsHTML = jobs.map(job => `
    <div class="job-card" onclick="navigateToJob(${job.id})">
      <h3 class="job-title">${escapeHtml(job.title)}</h3>
      <div class="job-meta">
        <span class="job-company">${escapeHtml(job.companyName)}</span>
        <span class="job-location">${escapeHtml(job.location)}</span>
      </div>
      <div class="job-date">Posted ${formatDate(job.firstSeenAt)}</div>
    </div>
  `).join('');
  
  jobsList.innerHTML = jobsHTML;
}

function updateJobsHeader(data, query) {
  const jobsTitle = document.getElementById('jobs-title');
  const jobsCount = document.getElementById('jobs-count');
  
  if (jobsTitle) {
    jobsTitle.textContent = query ? `Search Results for "${query}"` : 'All Jobs';
  }
  
  if (jobsCount) {
    jobsCount.textContent = `${data.total.toLocaleString()} jobs found`;
  }
}

function updatePagination(data, query) {
  const paginationContainer = document.getElementById('pagination');
  if (!paginationContainer) return;
  
  const { page, totalPages, total, pageSize } = data;
  
  if (totalPages <= 1) {
    paginationContainer.innerHTML = '';
    return;
  }
  
  let paginationHTML = '';
  
  // Previous button
  if (page > 1) {
    paginationHTML += `<a href="${buildJobsURL(query, page - 1, pageSize)}" class="pagination-btn">← Previous</a>`;
  } else {
    paginationHTML += `<span class="pagination-btn disabled">← Previous</span>`;
  }
  
  // Page numbers
  const startPage = Math.max(1, page - 2);
  const endPage = Math.min(totalPages, page + 2);
  
  if (startPage > 1) {
    paginationHTML += `<a href="${buildJobsURL(query, 1, pageSize)}" class="pagination-btn">1</a>`;
    if (startPage > 2) {
      paginationHTML += `<span class="pagination-info">...</span>`;
    }
  }
  
  for (let i = startPage; i <= endPage; i++) {
    if (i === page) {
      paginationHTML += `<span class="pagination-btn active">${i}</span>`;
    } else {
      paginationHTML += `<a href="${buildJobsURL(query, i, pageSize)}" class="pagination-btn">${i}</a>`;
    }
  }
  
  if (endPage < totalPages) {
    if (endPage < totalPages - 1) {
      paginationHTML += `<span class="pagination-info">...</span>`;
    }
    paginationHTML += `<a href="${buildJobsURL(query, totalPages, pageSize)}" class="pagination-btn">${totalPages}</a>`;
  }
  
  // Next button
  if (page < totalPages) {
    paginationHTML += `<a href="${buildJobsURL(query, page + 1, pageSize)}" class="pagination-btn">Next →</a>`;
  } else {
    paginationHTML += `<span class="pagination-btn disabled">Next →</span>`;
  }
  
  // Page info
  const startItem = (page - 1) * pageSize + 1;
  const endItem = Math.min(page * pageSize, total);
  paginationHTML += `<div class="pagination-info">Showing ${startItem}-${endItem} of ${total.toLocaleString()}</div>`;
  
  paginationContainer.innerHTML = paginationHTML;
}

function buildJobsURL(query, page, pageSize) {
  const params = new URLSearchParams();
  if (query) params.set('q', query);
  if (page > 1) params.set('page', page.toString());
  if (pageSize !== 20) params.set('page_size', pageSize.toString());
  
  const paramString = params.toString();
  return `/jobs${paramString ? '?' + paramString : ''}`;
}

function initializePagination() {
  // Pagination is handled by URL changes, no additional initialization needed
}

function showJobsLoading() {
  const jobsList = document.getElementById('jobs-list');
  if (jobsList) {
    jobsList.innerHTML = '<div class="loading">Loading jobs...</div>';
  }
}

function showJobsError(message) {
  const jobsList = document.getElementById('jobs-list');
  if (jobsList) {
    jobsList.innerHTML = `<div class="error">${escapeHtml(message)}</div>`;
  }
}

// Job detail page functionality
function initializeJobDetailPage() {
  const jobId = getJobIdFromURL();
  if (jobId) {
    loadJobDetail(jobId);
  }
}

function getJobIdFromURL() {
  const path = window.location.pathname;
  const match = path.match(/\/jobs\/(\d+)/);
  return match ? parseInt(match[1]) : null;
}

async function loadJobDetail(jobId) {
  try {
    const response = await fetch(`/api/jobs/${jobId}`);
    const data = await response.json();
    
    if (data.success) {
      displayJobDetail(data.data);
    } else {
      showJobDetailError('Job not found');
    }
  } catch (error) {
    console.error('Error loading job detail:', error);
    showJobDetailError('Failed to load job details');
  }
}

function displayJobDetail(job) {
  document.title = `${job.title} - ${job.companyName} | Bobber`;
  
  const jobDetailContainer = document.getElementById('job-detail');
  if (!jobDetailContainer) return;
  
  jobDetailContainer.innerHTML = `
    <div class="job-detail-header">
      <h1 class="job-detail-title">${escapeHtml(job.title)}</h1>
      <div class="job-detail-meta">
        <span class="job-detail-company">${escapeHtml(job.companyName)}</span>
        <span class="job-detail-location">${escapeHtml(job.location)}</span>
      </div>
      <div class="job-detail-date">Posted ${formatDate(job.firstSeenAt)}</div>
    </div>
    <div class="job-description">
      ${job.description || '<p>No description available.</p>'}
    </div>
  `;
}

function showJobDetailError(message) {
  const jobDetailContainer = document.getElementById('job-detail');
  if (jobDetailContainer) {
    jobDetailContainer.innerHTML = `<div class="error">${escapeHtml(message)}</div>`;
  }
}

// Utility functions
function escapeHtml(text) {
  if (!text) return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function formatDate(dateString) {
  try {
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = Math.abs(now - date);
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    
    if (diffDays === 1) return 'yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.ceil(diffDays / 7)} weeks ago`;
    if (diffDays < 365) return `${Math.ceil(diffDays / 30)} months ago`;
    
    return date.toLocaleDateString();
  } catch (error) {
    return 'recently';
  }
}

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
  // Focus search on '/' key
  if (e.key === '/' && !e.ctrlKey && !e.metaKey) {
    const searchInput = document.getElementById('search-input');
    if (searchInput && document.activeElement !== searchInput) {
      e.preventDefault();
      searchInput.focus();
    }
  }
}); 