// ==========================================
// ZENOLANG DEMO - JAVASCRIPT
// ==========================================

// Simple notification system
console.log('ZenoLang Demo App Loaded');

// Form validation helper
document.addEventListener('DOMContentLoaded', function () {
    console.log('DOM loaded, forms:', document.querySelectorAll('form').length);

    // Add form validation feedback
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', function (e) {
            console.log('Form submit event triggered');
            const requiredFields = form.querySelectorAll('[required]');
            let isValid = true;

            requiredFields.forEach(field => {
                if (!field.value || !field.value.trim()) {
                    isValid = false;
                    field.style.borderColor = 'var(--danger)';
                    console.log('Invalid field:', field.name);
                } else {
                    field.style.borderColor = 'var(--border)';
                }
            });

            if (!isValid) {
                e.preventDefault();
                alert('Please fill in all required fields');
                console.log('Form submission prevented - invalid fields');
                return false;
            }

            console.log('Form is valid, submitting...');
        });
    });

    // File upload preview
    const fileInputs = document.querySelectorAll('input[type="file"]');
    fileInputs.forEach(input => {
        input.addEventListener('change', function (e) {
            const file = e.target.files[0];
            if (file) {
                console.log('File selected:', file.name);
            }
        });
    });
});

// SSE Connection for real-time notifications (if on dashboard)
if (window.location.pathname.includes('/dashboard')) {
    console.log('Dashboard detected - SSE notifications could be connected here');
    // Uncomment to enable SSE:
    // const eventSource = new EventSource('/tutorial/max/notifications/stream');
    // eventSource.addEventListener('notification', function(e) {
    //     console.log('New notification:', e.data);
    // });
}
