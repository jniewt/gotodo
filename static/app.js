import {ApiService} from './api-service.js';
import {ListManager} from './list-manager.js';
import {UIManager} from './ui-manager.js';

const apiService = new ApiService(); // Assume this is already defined
const listManager = new ListManager(apiService);


document.addEventListener('DOMContentLoaded', async () => {
    await listManager.initLists();
    const uiManager = new UIManager(listManager);
    uiManager.displayLists();
    uiManager.displayDefaultList();
});
