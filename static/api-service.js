export class ApiService {
    constructor(baseURL = '/api') {
        this.baseURL = baseURL;
    }

    async request(endpoint, { method = 'GET', body = null, headers = {} } = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const options = {
            method,
            headers: { 'Content-Type': 'application/json', ...headers },
            body: body ? JSON.stringify(body) : null,
        };

        try {
            const response = await fetch(url, options);
            if (!response.ok) {
                let errorBody;
                try {
                    errorBody = await response.json();
                } catch (jsonError) {
                    // If there's an error parsing the error body, use a generic message
                    throw new Error('Unknown API error and unable to parse error response');
                }
                throw new Error(errorBody.error || 'Unknown API error');
            }
            // Handle 204 No Content specifically by returning null or a similar indicator
            if (response.status === 204) {
                return null;
            }
            // For all other successful responses, parse and return the JSON.
            return await response.json();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    fetchAllLists() {
        return this.request('/list');
    }

    fetchList(listName) {
        return this.request(`/list/${encodeURIComponent(listName)}`);
    }

    createList(listName) {
        return this.request('/list', {
            method: 'POST',
            body: { name: listName }
        });
    }

    deleteList(listName) {
        return this.request(`/list/${encodeURIComponent(listName)}`, {
            method: 'DELETE'
        });
    }

    // Add more methods as needed...
}
