function copyLink() {
    const urlElement = document.getElementById('shortUrl');
    if (!urlElement) return;

    const url = urlElement.innerText;
    navigator.clipboard.writeText(url).then(() => {
        const hint = document.getElementById('copyHint');
        const oldText = hint.innerText;
        hint.innerText = "Copied!";
        hint.style.color = "var(--pico-primary)";
        setTimeout(() => {
            hint.innerText = oldText;
            hint.style.color = "";
        }, 2000);
    }).catch(err => {
        console.error('Copy failed:', err);
    });
}

function toggleAlias() {
    const container = document.getElementById('aliasContainer');
    const toggle = document.getElementById('aliasToggle');
    if (container && toggle) {
        container.style.display = toggle.checked ? 'block' : 'none';
        if (toggle.checked) {
            const aliasInput = document.getElementById('alias');
            if (aliasInput) aliasInput.focus();
        }
    }
}

function toggleTheme() {
    const html = document.documentElement;
    const current = html.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    html.setAttribute('data-theme', next);
    localStorage.setItem('theme', next);
    updateThemeIcon(next);
}

function updateThemeIcon(theme) {
    const icon = document.getElementById('themeIcon');
    if (!icon) return;

    if (theme === 'dark') {
        // Moon icon
        icon.innerHTML = `
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
        `;
    } else {
        // Sun icon  
        icon.innerHTML = `
            <circle cx="12" cy="12" r="5"></circle>
            <line x1="12" y1="1" x2="12" y2="3"></line>
            <line x1="12" y1="21" x2="12" y2="23"></line>
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
            <line x1="1" y1="12" x2="3" y2="12"></line>
            <line x1="21" y1="12" x2="23" y2="12"></line>
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
        `;
    }
}

// Initialize on Load
document.addEventListener('DOMContentLoaded', () => {
    // Theme setup
    const savedTheme = localStorage.getItem('theme') ||
        (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    document.documentElement.setAttribute('data-theme', savedTheme);
    updateThemeIcon(savedTheme);

    // Alias setup
    const aliasInput = document.getElementById('alias');
    const aliasToggle = document.getElementById('aliasToggle');
    if (aliasInput && aliasInput.value !== "" && aliasToggle) {
        aliasToggle.checked = true;
        toggleAlias();
    }
});
