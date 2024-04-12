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

    createList(listName, listColour) {
        const rgbColour = hexToRGB(listColour);
        return this.request('/list', {
            method: 'POST',
            body: { name: listName, colour: rgbColour }
        });
    }

    editList(listName, listColour) {
        const rgbColour = hexToRGB(listColour);
        return this.request(`/list/${encodeURIComponent(listName)}`, {
            method: 'PATCH',
            body: { name: listName, colour: rgbColour }
        });
    }

    deleteList(listName) {
        return this.request(`/list/${encodeURIComponent(listName)}`, {
            method: 'DELETE'
        });
    }

    createTask(listName, task) {
        return this.request(`/list/${encodeURIComponent(listName)}`, {
            method: 'POST',
            body: task
        });
    }

    deleteTask(taskId) {
        return this.request(`/items/${taskId}`, {
            method: 'DELETE'
        });
    }

    updateTask(taskId, task) {
        return this.request(`/items/${taskId}`, {
            method: 'PATCH',
            body: task
        });
    }
}

function hexToRGB(hex) {
    let r = parseInt(hex.slice(1, 3), 16);
    let g = parseInt(hex.slice(3, 5), 16);
    let b = parseInt(hex.slice(5, 7), 16);
    return { r, g, b };
}
