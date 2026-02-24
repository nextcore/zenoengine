// ZenoStudio — Global JS Helpers

// Highlight active nav link based on current path
document.addEventListener('DOMContentLoaded', function () {
    const path = window.location.pathname;
    document.querySelectorAll('.nav-item').forEach(function (link) {
        const href = link.getAttribute('href');
        if (href && path === href) {
            link.classList.add('active');
        }
    });
});
