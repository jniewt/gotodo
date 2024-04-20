import {TaskUIManager} from './task-ui-manager.js';

export class UIManager {
    constructor(listManager) {
        this.defaultListName = 'Soon';
        this.listManager = listManager;
        this.addListModal = new AddListModal(this.listManager, this.onListCreateAttempt.bind(this));
        this.editListModal = new EditListModal(this.listManager, this.onListEditAttempt.bind(this));
        this.taskManager = new TaskUIManager(this.listManager, {
            tasksDisplay: document.getElementById('tasksDisplay'),
            addTaskButtonContainer: document.getElementById('addTaskButtonContainer'),
        }, this.refreshTasksDisplay.bind(this), this.showAlert.bind(this));

        this.init();
    }

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

        let combinedLists = this.listManager.lists.map(l => ({ ...l, filtered: false }))
            .concat(this.listManager.filteredLists.map(l => ({ ...l, filtered: true })));


        combinedLists.forEach(list => this.createListElement(list));
    }

    displayDefaultList() {
        const defaultList = this.listManager.listByName(this.defaultListName) || this.listManager.filteredListByName(this.defaultListName);
        if (defaultList === null) {
            return;
        }
        this.listManager.getTasks(defaultList.name).then((tasks) => {
            this.taskManager.displayTasks(defaultList, tasks);
        }).catch((error) => {
            this.showAlert(`Failed to fetch tasks for default list: ${error.message}`);
        });
    }

    // Create a list element
    createListElement(list) {
        const listElement = document.createElement('a');
        listElement.classList.add('list-group-item', 'list-group-item-action', 'd-flex', 'justify-content-between', 'align-items-center');
        listElement.href = '#';

        let isFiltered = list.filtered || false;

        // Icon selection based on the list type
        const iconClass = isFiltered ? 'bi-filter' : 'bi-list-task';
        const icon = document.createElement('i');
        icon.classList.add('bi', iconClass, 'me-2'); // 'me-2' for margin
        icon.style.color = `rgb(${list.colour.r}, ${list.colour.g}, ${list.colour.b})`;

        listElement.appendChild(icon); // Append the icon to the list element

        // Text content
        const text = document.createElement('span');
        text.textContent = list.name;
        listElement.appendChild(text); // Append the text next to the icon

        var sidebarOffcanvas = document.getElementById('sidebar');
        listElement.onclick = () => {
            this.listManager.getTasks(list.name).then((tasks) => {
                this.taskManager.displayTasks(list, tasks);
                const sidebar = bootstrap.Offcanvas.getInstance(sidebarOffcanvas);
                if (sidebar !== null) {
                    sidebar.hide();
                }
            }).catch((error) => {
                this.showAlert(`Failed to fetch tasks for list: ${error.message}`);
            });
            return false; // Prevent default anchor action
        };

        const dropdown = this.createDropdown(list.name);
        listElement.appendChild(dropdown);

        // hide the dropdown for filtered lists for now, but so that the layout doesn't break
        if (isFiltered) {
            dropdown.style.visibility = 'hidden';
        }

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
            <li><a class="dropdown-item" href="#" data-action="edit">Edit</a></li>
            <li><a class="dropdown-item" href="#" data-action="delete">Delete</a></li>
        </ul>
    `;

        const editOption = dropdown.querySelector('[data-action="edit"]');
        const deleteOption = dropdown.querySelector('[data-action="delete"]');

        editOption.addEventListener('click', this.handleEditListClick.bind(this, listName));
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

    onListEditAttempt(listName, error) {
        if (error) {
            // If there's an error, show an error alert
            this.showAlert(`Failed to edit list: ${error.message}`, 'danger');
        } else {
            // On successful addition, update the display and show a success message
            this.displayLists(this.listManager.lists);
            // If the edited list is the one that is currently displayed, refresh the tasks display
            if (this.taskManager.currentList !== null) {

            }
            if (this.taskManager.currentList.name === listName) {
                let updatedList = this.listManager.listByName(listName);
                this.taskManager.displayTasks(updatedList, updatedList.items);
            } else if (this.taskManager.currentList.filtered) {
                let filteredList = this.taskManager.currentList;
                this.taskManager.displayTasks(filteredList, filteredList.items);
            }
            this.showAlert('List saved successfully!', 'success');
        }
    }

    // Handle the click event for deleting a list. event must be the last argument, as it's passed by the event listener,
    // which happens after the arguments from bind.
    handleDeleteListClick(listName, event) {
        event.stopPropagation(); // Prevent event bubbling
        const list = this.listManager.listByName(listName);
        if (list.items.length > 0) {
            if (!confirm(`The list "${listName}" has tasks. Do you still want to delete it?`)) {
                return; // Exit if user cancels
            }
        }
        try {
            this.listManager.deleteList(listName).then(() => {
                this.showAlert('List deleted successfully!', 'success');
                this.displayLists(); // Refresh list display
            }).catch((error) => {
                this.showAlert(`Failed to delete list: ${error.message}`, 'danger');
            });
        } catch (error) {
            this.showAlert(`Failed to retrieve list: ${error.message}`, 'danger');
        }
    }

    handleEditListClick(listName, event) {
        event.stopPropagation(); // Prevent event bubbling
        this.editListModal.show(listName);
    }

    async refreshTasksDisplay(list) {
        this.listManager.getTasks(list.name).then((tasks) => {
            this.taskManager.displayTasks(list, tasks);
        }).catch((error) => {
            this.showAlert(`Failed to fetch tasks for list: ${error.message}`);
        });
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
        const listColour = document.getElementById('listColourInput').value;

        if (listName) {
            try {
                await this.listManager.createList(listName, listColour);

                this.hide();
                listNameInput.value = ''; // Reset input after successful creation
                this.onListAdded(null); // Indicate success by passing null for the error
            } catch (error) {
                console.error('Failed to create list:', error);
                this.onListAdded(error); // Pass the error to the callback
                // TODO Show an error message on the modal itself
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

class EditListModal {
    constructor(listManager, onListEditedCallback) {
        this.listManager = listManager;
        this.onListEdited = onListEditedCallback; // Callback when a list is edited
        this.modalId = 'editListModal';
        this.setupEventListeners();
    }

    setupEventListeners() {
        const form = document.getElementById('editListForm');
        form.addEventListener('submit', this.handleSubmit.bind(this));
    }

    async handleSubmit(event) {
        event.preventDefault();
        const listNameInput = document.getElementById('listNameEditInput');
        const listName = listNameInput.value.trim();
        const listColour = document.getElementById('listColourEditInput').value;


            try {
                await this.listManager.editList(listName, listColour);
                this.hide();
                this.onListEdited(listName, null); // Indicate success by passing null for the error
            } catch (error) {
                console.error('Failed to change list:', error);
                this.onListEdited(listName, error); // Pass the error to the callback
                // TODO Show an error message on the modal itself
            }

    }

    show(listName) {
        const modalElement = new bootstrap.Modal(document.getElementById(this.modalId));

        let list = this.listManager.listByName(listName);
        // prefill form fields
        document.getElementById('listNameEditInput').value = list.name;
        // list name can't be edited for now, disable
        document.getElementById('listNameEditInput').disabled = true;

        document.getElementById('listColourEditInput').value = `#${rgbToHex(list.colour.r, list.colour.g, list.colour.b)}`;

        modalElement.show();
    }

    hide() {
        const modalElement = bootstrap.Modal.getInstance(document.getElementById(this.modalId));
        if (modalElement) {
            modalElement.hide();
        }
    }
}

function rgbToHex(r, g, b) {
    // Helper function to convert a single color component into a hexadecimal string
    const toHex = c => c.toString(16).padStart(2, '0');
    return toHex(r) + toHex(g) + toHex(b);
}

