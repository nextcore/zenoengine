
// ZenoJS HTTP Client with JWT Support

const TOKEN_KEY = 'zeno_auth_token';

export const http = {
    // Configuration
    baseURL: '',

    // Set Base URL
    setBaseURL(url) {
        this.baseURL = url;
    },

    // Token Management
    setToken(token) {
        localStorage.setItem(TOKEN_KEY, token);
    },

    getToken() {
        return localStorage.getItem(TOKEN_KEY);
    },

    removeToken() {
        localStorage.removeItem(TOKEN_KEY);
    },

    // Request Helper
    async request(method, url, data = null, headers = {}) {
        const token = this.getToken();

        const config = {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                ...headers
            }
        };

        if (token) {
            config.headers['Authorization'] = `Bearer ${token}`;
        }

        if (data) {
            config.body = JSON.stringify(data);
        }

        const fullUrl = url.startsWith('http') ? url : (this.baseURL + url);

        try {
            const response = await fetch(fullUrl, config);

            if (response.status === 401) {
                // Unauthorized
                this.removeToken();
                // Dispatch event for router/store to handle
                window.dispatchEvent(new CustomEvent('zeno:unauthorized'));
                throw new Error('Unauthorized');
            }

            if (!response.ok) {
                const errData = await response.json().catch(() => ({}));
                throw new Error(errData.message || `HTTP Error ${response.status}`);
            }

            // Return JSON data
            return await response.json();
        } catch (error) {
            console.error('[ZenoHTTP]', error);
            throw error;
        }
    },

    get(url, headers) { return this.request('GET', url, null, headers); },
    post(url, data, headers) { return this.request('POST', url, data, headers); },
    put(url, data, headers) { return this.request('PUT', url, data, headers); },
    delete(url, headers) { return this.request('DELETE', url, null, headers); }
};
