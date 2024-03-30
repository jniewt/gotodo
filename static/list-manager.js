export class ListManager {
    #lists = [];

    constructor(apiService) {
        this.apiService = apiService;
    }

    // Call this method initially to load lists when the application starts
    async initLists() {
        await this.#fetchAllLists();
    }

    get lists() {
        return this.#lists;
    }

    async #fetchAllLists() {
        try {
            const data = await this.apiService.fetchAllLists();
            this.#lists = data.lists;
        } catch (error) {
            console.error('Failed to fetch lists:', error);
            throw error; // Ensure UIManager can react to this error
        }
    }

    async getTasks(listName) {
        let data;
        try {
            data = await this.apiService.fetchList(listName);
        } catch (error) {
            console.error('Failed to fetch tasks:', error);
            throw error; // Rethrow to handle in caller
        }
        return data.list.items
    }

    async createList(listName) {
        try {
            await this.apiService.createList(listName);
            await this.#fetchAllLists(); // Update internal state with new list
        } catch (error) {
            console.error('Failed to create list:', error);
            throw error; // Rethrow to handle in caller
        }
    }

    async deleteList(listName) {
        try {
            await this.apiService.deleteList(listName);
            await this.#fetchAllLists(); // Update internal state after deletion
        } catch (error) {
            console.error('Failed to delete list:', error);
            throw error; // Rethrow to handle in caller
        }
    }

    async createTask(listName, task) {
        let data;
        try {
            data = await this.apiService.createTask(listName, task);
        } catch (error) {
            console.error('Failed to create task:', error);
            throw error;
        }
        return data.task
    }

    async updateTask(id, task) {
        let data;
        try {
            data = await this.apiService.updateTask(id, task);
        } catch (error) {
            console.error('Failed to update task:', error);
            throw error;
        }
        return data.task
    }

    async deleteTask(id) {
        try {
            await this.apiService.deleteTask(id);
        } catch (error) {
            console.error('Failed to delete task:', error);
            throw error;
        }
    }
}