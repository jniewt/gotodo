import { TaskUIManager } from './task-ui-manager.js';

export class UIManager {
    constructor(listManager) {
        this.listManager = listManager;
        this.addListModal = new AddListModal(this.listManager, this.onListCreateAttempt.bind(this));
        this.taskManager = new TaskUIManager(this.listManager, {
            tasksDisplay: document.getElementById('tasksDisplay'),
            addTaskButtonContainer: document.getElementById('addTaskButtonContainer'),
        }, this.refreshTasksDisplay.bind(this), this.showAlert.bind(this));

        this.init();
    }

    // Initialize the class
    init() {
        this.cacheDomElements();
        this.setupEventListeners();
    }

    // Cache DOM elements to improve performance
    cacheDomElements() {
        this.dom = {
            listsDisplay: document.getElementById('listsDisplay'),
            tasksDisplay: document.getElementById('tasksDisplay'),
            alertPlaceholder: document.getElementById('alertPlaceholder'), // Caching the alert placeholder
        };
    }

    // Set up all event listeners
    setupEventListeners() {

    }

    // Display lists on the UI
    displayLists() {
        this.dom.listsDisplay.innerHTML = ''; // Clear current lists
        let lists = this.listManager.lists;
        lists.forEach((listName) => this.createListElement(listName));
    }

    // Create a list element
    createListElement(listName) {
        const listElement = document.createElement('a');
        listElement.classList.add('list-group-item', 'list-group-item-action', 'd-flex', 'justify-content-between', 'align-items-center');
        listElement.textContent = listName;
        listElement.href = '#';

        listElement.onclick = () => {
            this.listManager.getTasks(listName).then((tasks) => {
                this.taskManager.displayTasks(listName, tasks);
            }).catch((error) => {
                this.showAlert(`Failed to fetch tasks for list: ${error.message}`);
            });
            return false; // Prevent default anchor action
        };

        const dropdown = this.createDropdown(listName);
        listElement.appendChild(dropdown);
        this.dom.listsDisplay.appendChild(listElement);
    }

    // Create the dropdown menu for a list.
    createDropdown(listName) {
        let dropdown = document.createElement('div');
        dropdown.classList.add('dropdown');
        dropdown.innerHTML = `
      <a class="text-secondary" href="#" role="button" data-bs-toggle="dropdown" aria-expanded="false">
        <i class="bi bi-three-dots-vertical"></i>
      </a>
      <ul class="dropdown-menu" aria-labelledby="dropdownMenuButton">
        <li><a class="dropdown-item" href="#">Delete</a></li>
      </ul>
    `;

        const deleteOption = dropdown.querySelector('.dropdown-item');
        deleteOption.addEventListener('click', this.handleDeleteListClick.bind(this, listName));

        // Prevent list selection when interacting with the dropdown
        dropdown.addEventListener('click', (event) => event.stopPropagation());

        return dropdown;
    }

    onListCreateAttempt(error) {
        if (error) {
            // If there's an error, show an error alert
            this.showAlert(`Failed to create list: ${error.message}`, 'danger');
        } else {
            // On successful addition, update the display and show a success message
            this.displayLists(this.listManager.lists);
            this.showAlert('List created successfully!', 'success');
        }
    }

    // Handle the click event for deleting a list. event must be the last argument, as it's passed by the event listener,
    // which happens after the arguments from bind.
    handleDeleteListClick(listName, event) {
        event.stopPropagation(); // Prevent event bubbling
        this.listManager.deleteList(listName).then(() => {
            this.showAlert('List deleted successfully!', 'success');
            this.displayLists(); // Refresh list display
        }).catch((error) => {
            this.showAlert(`Failed to delete list: ${error.message}`, 'danger');
        });
    }

    async refreshTasksDisplay(listName) {
        const tasks = await this.listManager.getTasks(listName)
        this.taskManager.displayTasks(listName, tasks);
    }

    showAlert(message, type = 'danger') {
        // Using the cached alertPlaceholder for performance
        const alertPlaceholder = this.dom.alertPlaceholder;

        // alertPlaceholder.innerHTML = ''; // This line clears previous alerts. Remove if you prefer to stack alerts.

        const alertWrapper = document.createElement('div');
        alertWrapper.innerHTML = [
            `<div class="alert alert-${type} alert-dismissible fade show" role="alert">`,
            `   <span>${message}</span>`,
            '   <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>',
            '</div>'
        ].join('');

        alertPlaceholder.append(alertWrapper);

        // Optional: Automatically dismiss the alert after a certain time
        setTimeout(() => {
            alertWrapper.remove();
        }, 5000); // Adjust timing as needed

    }
}

class AddListModal {
    constructor(listManager, onListAddedCallback) {
        this.listManager = listManager;
        this.onListAdded = onListAddedCallback; // Callback when a list is added
        this.modalId = 'addListModal';
        this.initModal();
    }

    initModal() {
        const modalHTML = `
      <div class="modal fade" id="${this.modalId}" tabindex="-1" aria-labelledby="${this.modalId}Label" aria-hidden="true">
        <div class="modal-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title" id="${this.modalId}Label">New List</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
            </div>
            <div class="modal-body">
              <form id="addListForm">
                <div class="mb-3">
                  <label for="listNameInput" class="form-label">List Name</label>
                  <input type="text" class="form-control" id="listNameInput" required>
                </div>
                <button type="submit" class="btn btn-primary">Create List</button>
              </form>
            </div>
          </div>
        </div>
      </div>
    `;

        document.body.insertAdjacentHTML('beforeend', modalHTML);
        this.setupEventListeners();
    }

    setupEventListeners() {
        const form = document.getElementById('addListForm');
        form.addEventListener('submit', this.handleSubmit.bind(this));
    }

    async handleSubmit(event) {
        event.preventDefault();
        const listNameInput = document.getElementById('listNameInput');
        const listName = listNameInput.value.trim();

        if (listName) {
            try {
                await this.listManager.createList(listName);

                this.hide();
                listNameInput.value = ''; // Reset input after successful creation
                this.onListAdded(null); // Indicate success by passing null for the error
            } catch (error) {
                console.error('Failed to create list:', error);
                this.onListAdded(error); // Pass the error to the callback
            }
        }
    }

    show() {
        const modalElement = new bootstrap.Modal(document.getElementById(this.modalId));
        modalElement.show();
    }

    hide() {
        const modalElement = bootstrap.Modal.getInstance(document.getElementById(this.modalId));
        if (modalElement) {
            modalElement.hide();
        }
    }
}