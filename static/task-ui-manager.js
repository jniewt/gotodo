import { sortByTitleThenDone, sortTasks } from './sort-tasks.js';
import {formatDate, formatDateHuman} from "./format-date.js";

export class TaskUIManager {
    constructor(listManager, domElements, onTaskListChange = () => {}, showAlert = (msg, type) => {console.log(msg)}) {
        this.listManager = listManager;
        this.addTaskModal = new AddTaskModal(listManager, onTaskListChange, showAlert);
        this.taskDetailsModal = new TaskDetailsModal();
        this.dom = domElements;
        this.currentListName = '';
        this.showAlert = showAlert;
        this.onTaskListChange = onTaskListChange; // Callback when task list changes
        this.initContextMenu();
        this.initAddTaskButton();
    }

    displayTasks(listName, tasks) {
        this.currentListName = listName;
        const { tasksDisplay } = this.dom;
        tasksDisplay.innerHTML = `<h3>${listName}</h3>`;

        if (!tasks || tasks.length === 0) {
            tasksDisplay.innerHTML += '<p>No tasks in this list.</p>';
            return;
        }

        let sorted = sortTasks(tasks, sortByTitleThenDone);

        const taskList = document.createElement('ul');
        taskList.className = 'list-group';
        sorted.forEach(task => taskList.appendChild(this.createTaskElement(task)));
        tasksDisplay.appendChild(taskList);
    }

    createTaskElement(task) {
        const itemEl = document.createElement('li');
        itemEl.className = 'list-group-item d-flex align-items-center';

        const checkbox = this.createCheckbox(task);
        const titleSpan = this.createTitleSpan(task);

        checkbox.addEventListener('change', () => this.handleTaskStatusChange(task, checkbox, titleSpan));

        // Prevent event from triggering when clicking on the checkbox
        checkbox.onclick = event => event.stopPropagation();
        itemEl.addEventListener('click', () => {
            this.taskDetailsModal.setCurrentTask(task); // Set the current task
            this.taskDetailsModal.show(); // Show the modal with task details
        });

        itemEl.addEventListener('contextmenu', (event) => {
            event.preventDefault(); // Prevent the default context menu from appearing
            this.currentRightClickedTaskId = task.id; // Store the current task id for deletion
            this.contextMenu.style.top = `${event.clientY}px`;
            this.contextMenu.style.left = `${event.clientX}px`;
            this.contextMenu.style.display = 'block';
        });

        itemEl.append(checkbox, titleSpan);
        return itemEl;
    }

    createCheckbox(task) {
        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.className = 'me-2';
        checkbox.checked = task.done;
        return checkbox;
    }

    createTitleSpan(task) {
        const titleSpan = document.createElement('span');
        titleSpan.textContent = task.title;
        this.updateTitleSpanAppearance(titleSpan, task.done);
        return titleSpan;
    }

    updateTitleSpanAppearance(titleSpan, isDone) {
        titleSpan.classList.toggle('text-decoration-line-through', isDone);
        titleSpan.classList.toggle('text-muted', isDone);
    }

    async handleTaskStatusChange(task, checkbox, titleSpan) {
        this.updateTitleSpanAppearance(titleSpan, checkbox.checked);
        try {
            await this.listManager.updateTask(task.id, { done: checkbox.checked });
            this.onTaskListChange(this.currentListName); // Notify the parent UI manager of the change
        } catch (error) {
            console.error('Error updating task status:', error);
            // Error handling logic can be implemented here.
        }
    }

    handleDeleteTask(taskId) {
        this.listManager.deleteTask(taskId).then(() => {
            this.onTaskListChange(this.currentListName);
            this.showAlert('Task deleted', 'success');
        }).catch(error => {
            // Log the error to the console and display an error message to the user
            console.error('Error deleting task:', error);
            // Use the showAlert method to display the error message on the main page
            this.showAlert(`Failed to delete the task: ${error.message}`, 'danger');
        });
    }


    initContextMenu() {
        this.contextMenu = document.createElement('div');
        this.contextMenu.innerHTML = `<ul class="list-group">
            <li class="list-group-item list-group-item-action" id="delete-task">Delete</li>
        </ul>`;
        this.contextMenu.style.position = 'absolute';
        this.contextMenu.style.display = 'none';
        document.body.appendChild(this.contextMenu);

        document.addEventListener('click', () => {
            this.contextMenu.style.display = 'none';
        });

        this.contextMenu.querySelector('#delete-task').addEventListener('click', () => {
            // Call the delete handler with the current task's id
            if(this.currentRightClickedTaskId) {
                this.handleDeleteTask(this.currentRightClickedTaskId);
                this.currentRightClickedTaskId = null; // Reset the current task id after deletion
            }
            this.contextMenu.style.display = 'none';
        });
    }

    initAddTaskButton() {
        const addTaskButton = document.createElement('button');
        addTaskButton.textContent = 'Add Task';
        addTaskButton.className = 'btn btn-primary';
        addTaskButton.addEventListener('click', () => {
            this.addTaskModal.setCurrentList(this.currentListName); // Update current list before showing
            this.addTaskModal.show();
        });

        this.dom.addTaskButtonContainer.appendChild(addTaskButton);
    }
}

class AddTaskModal {
    constructor(listManager, onTaskAddedCallback, showAlert) {
        this.listManager = listManager;
        this.currentList = '';
        this.onTaskAdded = onTaskAddedCallback;
        this.modalId = 'addTaskModal';
        this.showAlertOutside = showAlert;
        this.initModal();
    }

    initModal() {
        const modalHTML = `
            <div class="modal fade" id="${this.modalId}" tabindex="-1">
                <div class="modal-dialog modal-dialog-centered">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title" id="addItemModalLabel">Add New Task</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                        </div>
                        <div class="modal-body">
                            <form id="addItemForm"> 
                                <!-- Error message section, initially hidden -->
                                <div class="alert alert-danger d-none" id="formErrorAlert"></div>
                                
                                <div class="mb-3">
                                    <label for="taskTitleInput" class="form-label">Title</label>
                                    <input type="text" class="form-control" id="taskTitleInput" required>
                                </div>
                                <div class="mb-3">
                                    <label for="taskListDropdown" class="form-label">List</label>
                                    <select class="form-select" id="taskListDropdown" required>
                                        <!-- Options will be dynamically added here -->
                                    </select>
                                </div>
                                <div class="mb-3">
                                    <label for="dueDateTypeSelect" class="form-label">Due Date</label>
                                    <select class="form-select" id="dueDateTypeSelect">
                                        <option value="none" selected>None</option>
                                        <option value="dueOn">Due On</option>
                                        <option value="dueBy">Due By</option>
                                    </select>
                                </div>
                                <div class="mb-3 d-none" id="dateTimeOptions">
                                    <div class="mb-3 form-check">
                                        <input type="checkbox" class="form-check-input" id="taskAllDayInput" checked>
                                        <label class="form-check-label" for="taskAllDayInput">All Day</label>
                                    </div>
                                    <div class="mb-3">
                                        <label for="taskDueDateTime" class="form-label">Date</label>
                                        <input type="date" class="form-control" id="taskDueDateTime">
                                    </div>
                                </div>
                            </form>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                            <button type="button" class="btn btn-primary" id="saveTaskButton">OK</button>
                        </div>
                    </div>
                </div>
            </div>
        `;

        document.body.insertAdjacentHTML('beforeend', modalHTML);

        this.setupEventListeners();
    }

    setCurrentList(currentListName) {
        this.currentList = currentListName;
        this.populateListDropdown(); // Ensure dropdown is updated with the current selection
    }

    async populateListDropdown() {
        const dropdown = document.getElementById('taskListDropdown');
        dropdown.innerHTML = ''; // Clear existing options
        const lists = this.listManager.lists;
        lists.forEach(list => {
            const option = new Option(list, list, list === this.currentList, list === this.currentList);
            dropdown.add(option);
        });
    }

    setupEventListeners() {
        const saveButton = document.getElementById('saveTaskButton');
        const dueDateTypeSelect = document.getElementById('dueDateTypeSelect');
        const modalElement = document.getElementById(this.modalId);

        // Add an event listener for when the modal is fully hidden
        modalElement.addEventListener('hidden.bs.modal', () => {
            this.hide(); // Call your custom hide method
        });

        // Handle form submission
        saveButton.addEventListener('click', () => this.handleSubmit());

        // Show date/time options based on due date type selection
        dueDateTypeSelect.addEventListener('change', function() {
            const dateTimeOptions = document.getElementById('dateTimeOptions');
            const dueDateTimeInput = document.getElementById('taskDueDateTime');
            const allDayCheckbox = document.getElementById('taskAllDayInput');

            if (this.value === 'none') {
                dateTimeOptions.classList.add('d-none');
            } else {
                // Remove 'd-none' to show the options
                dateTimeOptions.classList.remove('d-none');
                // Check the 'All Day' checkbox by default
                allDayCheckbox.checked = true;

                // Determine today's date
                const today = new Date();
                const todayFormattedDate = today.toISOString().split('T')[0];
                const todayFormattedDateTime = todayFormattedDate + 'T' + today.toTimeString().split(' ')[0];
                // Set the input value to today, adjusting format based on 'All Day'
                dueDateTimeInput.value = allDayCheckbox.checked ? todayFormattedDate : todayFormattedDateTime;
                // Since 'All Day' is checked by default, set type to 'date'
                dueDateTimeInput.type = 'date';
            }
        });

        document.getElementById('taskAllDayInput').addEventListener('change', function() {
            const dueDateTimeInput = document.getElementById('taskDueDateTime');
            if (this.checked) {
                // Change the input type to 'date', removing the time part but keeping the date
                const currentValue = dueDateTimeInput.value;
                if (currentValue) {
                    const datePart = currentValue.includes('T') ? currentValue.split('T')[0] : currentValue;
                    dueDateTimeInput.type = 'date';
                    dueDateTimeInput.value = datePart; // Keep the previously selected date
                } else {
                    // If there was no previous value, simply switch to date input
                    dueDateTimeInput.type = 'date';
                }
            } else {
                // When unchecking 'All Day', enable time selection without resetting the date
                const currentValue = dueDateTimeInput.value;
                dueDateTimeInput.type = 'datetime-local';
                if (currentValue && !currentValue.includes('T')) {
                    // If there's already a date but no time, append a default time part to it
                    // This ensures the input value format matches 'datetime-local' requirements
                    dueDateTimeInput.value = `${currentValue}T09:00`; // Default to 9 AM
                }
                // Note: If there was already a datetime value, changing the input type back to 'datetime-local'
                // will naturally preserve it, so there's no need to explicitly set it again.
            }
        });
    }

    async handleSubmit() {
        // Validate the form
        const form = document.getElementById('addItemForm');
        if (!form.checkValidity()) {
            form.classList.add('was-validated'); // Bootstrap 5 validation
            form.reportValidity();
            return;
        }

        const titleInput = document.getElementById('taskTitleInput');
        const listDropdown = document.getElementById('taskListDropdown');
        const dueDateTypeSelect = document.getElementById('dueDateTypeSelect');
        const allDayCheckbox = document.getElementById('taskAllDayInput');
        const dueDateTimeInput = document.getElementById('taskDueDateTime');

        const title = titleInput.value.trim();
        const listName = listDropdown.value;
        const dueDateType = dueDateTypeSelect.value;
        const isAllDay = allDayCheckbox.checked;
        const dueDateTime = dueDateTimeInput.value;

        let requestPayload = {
            title: title,
        };

        if (dueDateType === 'dueOn') {
            requestPayload.due_on = dueDateTime;
            requestPayload.all_day = isAllDay;
        } else if (dueDateType === 'dueBy') {
            requestPayload.due_by = dueDateTime;
            requestPayload.all_day = isAllDay;
        }

        try {
            console.log('Adding task to list:', listName, requestPayload)
            await this.listManager.createTask(listName, requestPayload);
            this.showAlert('Task added successfully', 'success');
            this.hide(); // Hide the modal
            this.onTaskAdded(listName); // Trigger the callback on added task
            this.showAlertOutside('Task added', 'success');
        } catch (error) {
            // TODO doesn't make sense since the alert appears behind the modal, either close the modal or show the alert in the modal
            console.error('Error adding task:', error);
            this.showAlert(`Failed to add the task: ${error.message}`);
        }
    }

    show() {
        const form = document.getElementById('addItemForm');
        form.reset(); // Reset the form to clear any previous values
        this.populateListDropdown(); // Ensure the dropdown is up-to-date
        const modalElement = new bootstrap.Modal(document.getElementById(this.modalId));
        modalElement.show();
    }

    hide() {
        const modalElement = bootstrap.Modal.getInstance(document.getElementById(this.modalId));
        const form = document.getElementById('addItemForm');
        if (modalElement) {
            modalElement.hide();
            form.reset();
            document.getElementById('taskTitleInput').value = '';
            form.classList.remove('was-validated'); // Remove validation class to reset the form state
            // Also hide the alert box when modal is closed or the task is successfully added
            const alertBox = document.getElementById('formErrorAlert');
            alertBox.classList.add('d-none');
            alertBox.textContent = ''; // Clear the error message
        }
    }

    showAlert(message, type) {
        const alertBox = document.getElementById('formErrorAlert');
        alertBox.textContent = message;
        alertBox.classList.remove('d-none'); // Show the alert box
        alertBox.classList.add(`alert-${type}`); // Use the 'type' to add specific styling, e.g., 'alert-danger' for errors
    }
}

class TaskDetailsModal {
    constructor() {
        this.currentTask = null;
        this.modalId = 'taskDetailsModal';
        this.initModal();
    }

    initModal() {
        const modalHTML = `
            <div class="modal fade" id="taskDetailsModal" tabindex="-1" aria-labelledby="taskDetailsModalLabel" aria-hidden="true">
                <div class="modal-dialog modal-dialog-centered">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title" id="taskDetailsModalLabel">Task Details</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                        </div>
                        <div class="modal-body">
                            <dl class="row">
                                <dt class="col-sm-4">Title:</dt>
                                <dd class="col-sm-8" id="taskTitle"></dd>

                                <dt class="col-sm-4">List:</dt>
                                <dd class="col-sm-8" id="taskList"></dd>

                                <dt class="col-sm-4">Status:</dt>
                                <dd class="col-sm-8" id="taskStatus"></dd>

                                <dt class="col-sm-4">All day:</dt>
                                <dd class="col-sm-8" id="taskAllDay"></dd>

                                <dt class="col-sm-4" id="taskDueOnLabel">Due On:</dt>
                                <dd class="col-sm-8" id="taskDueOn"></dd>

                                <dt class="col-sm-4" id="taskDueByLabel">Due By:</dt>
                                <dd class="col-sm-8" id="taskDueBy"></dd>

                                <dt class="col-sm-4">Created:</dt>
                                <dd class="col-sm-8" id="taskCreated"></dd>

                                <dt class="col-sm-4" id="taskDoneOnLabel">Done On:</dt>
                                <dd class="col-sm-8" id="taskDoneOn"></dd>
                            </dl>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                        </div>
                    </div>
                </div>
            </div>
        `;
        document.body.insertAdjacentHTML('beforeend', modalHTML);
        this.modalElement = new bootstrap.Modal(document.getElementById(this.modalId));
        this.setupEventListeners();
    }

    setupEventListeners() {
        const modalElement = document.getElementById(this.modalId);

        // Add an event listener for when the modal is fully hidden
        modalElement.addEventListener('hidden.bs.modal', () => {
            this.hide();
        });
    }

    setCurrentTask(task) {
        this.currentTask = task;
        this.populateModalFields(task);
    }

    populateModalFields(task) {
        document.getElementById('taskTitle').textContent = task.title;
        document.getElementById('taskList').textContent = task.list;
        document.getElementById('taskStatus').textContent = task.done ? 'Completed' : 'Pending';
        document.getElementById('taskCreated').textContent = formatDate(task.created);

        const allDay = task.all_day;
        document.getElementById('taskAllDay').textContent = allDay ? 'Yes' : 'No';

        // Conditionally populate and display the "Due On" and "Due By" fields

        const dueOn = task.due_on;
        const dueBy = task.due_by;

        document.getElementById('taskDueOn').textContent = dueOn ? formatDateHuman(dueOn, allDay) : '';


        document.getElementById('taskDueBy').textContent = dueBy ? formatDateHuman(dueBy, allDay) : '';


        // Conditionally populate and display the "Done On" field
        const doneOn = task.done_on;

        const isTaskDone = task.done === 'true';
        document.getElementById('taskDoneOn').textContent = isTaskDone ? formatDate(doneOn) : '';


        // Adjust visibility of date labels based on data
        this.adjustDateVisibility(task);
    }

    adjustDateVisibility(task) {
        document.getElementById('taskDueOn').style.display = task.due_on ? 'block' : 'none';
        document.getElementById('taskDueOnLabel').style.display = task.due_on ? 'block' : 'none';
        document.getElementById('taskDueBy').style.display = task.due_by ? 'block' : 'none';
        document.getElementById('taskDueByLabel').style.display = task.due_by ? 'block' : 'none';
        const isTaskDone = task.done === 'true';
        document.getElementById('taskDoneOnLabel').style.display = isTaskDone ? 'block' : 'none';
        document.getElementById('taskDoneOn').style.display = isTaskDone ? 'block' : 'none';
    }

    show() {
        this.modalElement.show();
    }

    hide() {
        this.modalElement.hide();
    }
}