import {sortByDone, sortByTitle, sortTasks, sortByPriority, sortByDueDate} from './sort-tasks.js';
import {formatDate, formatDateForm, formatDateHuman} from "./format-date.js";

export class TaskUIManager {
    listManager;
    addTaskModal;
    // TODO why does this have to be initialized here, and addTaskModal can be initialized in the constructor?
    // otherwise the task title field is not getting populated
    taskDetailsModal;
    dom;
    currentList = null;
    showAlert;
    onTaskListChange;

    constructor(listManager, domElements, onTaskListChange = () => {}, showAlert = (msg, type) => {console.log(msg)}) {
        this.listManager = listManager;
        this.dom = domElements;
        this.showAlert = showAlert;
        this.onTaskListChange = onTaskListChange;
        this.addTaskModal = new AddTaskModal(listManager, onTaskListChange, showAlert);
        this.taskDetailsModal = new TaskDetailsModal(this.listManager, this.showAlert);
        this.taskDetailsModal.setListManager(listManager);
        this.taskDetailsModal.setTaskUpdateCallback(onTaskListChange);

        this.init();
    }

    init() {
        this.initContextMenu();
        this.initAddTaskButton();
    }

    displayTasks(list, tasks) {
        this.currentList = list;
        this.clearTasksDisplay();
        this.setListNameHeader(list.name);
        const mainPanel = document.getElementById('mainPanel');
        applyListColour(list, mainPanel);

        if (!tasks || tasks.length === 0) {
            this.showNoTasksMessage();
            return;
        }

        this.displaySortedTasks(tasks, list.filtered);
    }

    clearTasksDisplay() {
        const { tasksDisplay } = this.dom;
        tasksDisplay.innerHTML = '';
    }

    setListNameHeader(listName) {
        const header = document.getElementById('listHeader');
        header.textContent = listName;
    }

    showNoTasksMessage() {
        this.dom.tasksDisplay.innerHTML += '<p>No tasks in this list.</p>';
    }

    displaySortedTasks(tasks, filtered=false) {
        let sorted = sortTasks(tasks, [sortByDone, sortByDueDate, sortByPriority, sortByTitle]);
        const taskList = document.createElement('ul');
        taskList.className = 'list-group';
        sorted.forEach(task => taskList.appendChild(this.createTaskElement(task, filtered)));
        this.dom.tasksDisplay.appendChild(taskList);
    }

    createTaskElement(task, filtered=false) {
        const itemEl = document.createElement('a');
        itemEl.href = '#';
        itemEl.className = 'list-group-item d-flex align-items-center';

        // if list is filtered, add a color to the task
        if (filtered) {
            let list = this.listManager.listByName(task.list);
            itemEl.style.backgroundColor = `rgba(${list.colour.r}, ${list.colour.g}, ${list.colour.b}, 0.1)`;
        }

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

        const contentContainer = document.createElement('div');
        contentContainer.className = 'd-flex align-items-center flex-grow-1';

        contentContainer.append(checkbox)
        // Show priority icon if task is not normal priority
        if (task.priority !== 0) {
            const priorityIcon = document.createElement('i');
            let iconClass, iconColor;
            switch (task.priority) {
                case -2:
                    iconClass = 'bi-chevron-double-down';
                    iconColor = 'text-success';
                    titleSpan.classList.add('fw-lighter')
                    break;
                case -1:
                    iconClass = 'bi-chevron-down';
                    iconColor = 'text-success-emphasis';
                    break;
                case 1:
                    iconClass = 'bi-chevron-up';
                    iconColor = 'text-warning';
                    titleSpan.classList.add('fw-medium')
                    break;
                case 2:
                    iconClass = 'bi-chevron-double-up';
                    iconColor = 'text-danger';
                    titleSpan.classList.add('fw-semibold')
                    break;
            }
            priorityIcon.className = `bi ${iconClass} ${iconColor} me-1`;

            contentContainer.append(priorityIcon);
        }
        contentContainer.append(titleSpan)

        itemEl.append(contentContainer);

        // check if the task has a due date and add the due date icon and text
        if (task.due_type === "due_by" || task.due_type === "due_on" || task.done) {
            const dueDateInfo = document.createElement('span');
            let dateText, iconClass, textStyle;

            if (task.done) {
                // For completed tasks, show the completion date with specific styling, all_day is not important anymore
                dateText = formatDateHuman(task.done_on);
                iconClass = 'bi-calendar-check'; // An icon indicating completion
                textStyle = 'color: grey; text-decoration: line-through;';
            } else {
                // Handle due and overdue tasks
                const dueDate = task.due;
                const overdue = isOverdue(new Date(dueDate), new Date(), task.all_day);
                dateText = formatDateHuman(dueDate, task.all_day);
                iconClass = task.due_type === "due_by" ? 'bi-calendar-range' : 'bi-calendar';
                textStyle = overdue ? 'color: red;' : 'color: inherit;';
            }

            dueDateInfo.innerHTML = `
        <i class="bi ${iconClass} me-2" style="font-size: 0.75rem; margin-right: 4px;"></i>
        <span style="display: inline-block; width: 70px; ${textStyle}">${dateText}</span>
    `;
            // Adjusted font size for the entire dueDateInfo, including the icon and text
            dueDateInfo.style.fontSize = '0.55rem';
            dueDateInfo.classList.add('align-self-end');

            const wrapper = document.createElement('div');
            wrapper.className = 'd-flex justify-content-between align-items-center flex-grow-1';
            wrapper.appendChild(contentContainer); // Add the checkbox and title
            wrapper.appendChild(dueDateInfo); // Add the due date or done date info

            itemEl.appendChild(wrapper);
        }

        return itemEl;
    }

    createCheckbox({ done }) {
        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.className = 'me-2';
        checkbox.checked = done;
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
            this.onTaskListChange(this.currentList); // Notify the parent UI manager of the change
        } catch (error) {
            console.error('Error updating task status:', error);
            // Error handling logic can be implemented here.
        }
    }

    handleDeleteTask(taskId) {
        this.listManager.deleteTask(taskId).then(() => {
            this.onTaskListChange(this.currentList);
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
            <a class="list-group-item list-group-item-action" id="delete-task" href="#">Delete</a>
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
        addTaskButton.textContent = '+';
        addTaskButton.className = 'btn btn-warning rounded-circle btn-lg m-2 mt-0';
        addTaskButton.setAttribute('title', 'Add Task'); // Tooltip text
        addTaskButton.setAttribute('data-bs-toggle', 'tooltip'); // Bootstrap tooltip
        addTaskButton.setAttribute('data-bs-placement', 'bottom'); // Tooltip position
        addTaskButton.addEventListener('click', () => {
            this.addTaskModal.setCurrentList(this.currentList); // Update current list before showing
            this.addTaskModal.show();
        });

        this.dom.addTaskButtonContainer.appendChild(addTaskButton);
    }
}

class AddTaskModal {
    constructor(listManager, onTaskAddedCallback, showAlert) {
        this.listManager = listManager;
        this.currentList = null;
        this.onTaskAdded = onTaskAddedCallback;
        this.modalId = 'addTaskModal';
        this.showAlertOutside = showAlert;
        this.setupEventListeners();
    }

    setCurrentList(list) {
        this.currentList = list;
        this.populateListDropdown().then(r => {}); // Ensure dropdown is updated with the current selection
    }

    async populateListDropdown() {
        const dropdown = document.getElementById('taskListDropdown');
        dropdown.innerHTML = ''; // Clear existing options
        const lists = this.listManager.lists;
        lists.forEach(list => {
            const isSelected = this.currentList && list.name === this.currentList.name;
            const option = new Option(list.name, list.name, isSelected, isSelected);
            dropdown.add(option);
        });
    }

    setupEventListeners() {
        const modalElement = document.getElementById(this.modalId);

        // focus on the title input field when the modal is shown
        modalElement.addEventListener('shown.bs.modal', () => {
            document.getElementById('taskTitleInput').focus();
        });

        const saveButton = document.getElementById('saveTaskButton');

        // Add an event listener for when the modal is fully hidden
        modalElement.addEventListener('hidden.bs.modal', () => {
            this.hide(); // Call your custom hide method
        });

        // Handle form submission
        saveButton.addEventListener('click', () => this.handleSubmit());
        modalElement.addEventListener('keydown', function(event) {
            if (event.key === 'Enter') {
                event.preventDefault();
                saveButton.click();
            }
        })

        // Show date/time options based on due date type selection
        handleDueTypeChange(
            document.getElementById('dueDateTypeSelect'),
            'dateTimeOptions',
            'timeOptions',
            'taskDueDate',
            'taskAllDayInput'
        );

        document.getElementById('taskAllDayInput').addEventListener('change', function() {
            const timeOptions = document.getElementById('timeOptions');
            if (this.checked) {
                // hide the time input
                timeOptions.classList.add('d-none');
            } else {
                // show the time input
                timeOptions.classList.remove('d-none');
                // default to 9 AM
                document.getElementById('taskDueTime').value = '09:00';
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
        const dueDateInput = document.getElementById('taskDueDate');
        const dueTimeInput = document.getElementById('taskDueTime');
        const prioritySelect = document.getElementById('prioritySelect');

        const title = titleInput.value.trim();
        const listName = listDropdown.value;
        const priority = prioritySelect.value;
        const dueDateType = dueDateTypeSelect.value;
        const isAllDay = allDayCheckbox.checked;
        const dueDate = dueDateInput.value;
        const dueTime = dueTimeInput.value;
        // format a date string in the format 'YYYY-MM-DDTHH:MM:SS'
        const dueDateTime = isAllDay ? `${dueDate}` : `${dueDate}T${dueTime}`;

        let requestPayload = {
            title: title,
            due_type: dueDateType,
            priority: parseInt(priority,10),
        };

        if (dueDateType !== 'none') {
            requestPayload.due = dueDateTime;
            requestPayload.all_day = isAllDay;
        }

        try {
            console.log('Adding task to list:', listName, requestPayload)
            await this.listManager.createTask(listName, requestPayload);
            this.hide(); // Hide the modal
            // If current list is filtered, update it, otherwise show the target list
            let targetList = this.currentList && this.currentList.filtered ? this.currentList : this.listManager.listByName(listName);
            this.onTaskAdded(targetList); // Trigger the callback on added task
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
        document.getElementById('taskTitleInput').value = '';
        document.getElementById('dueDateTypeSelect').value = 'none'; // Reset due date type
        document.getElementById('dateTimeOptions').classList.add('d-none');
        document.getElementById('taskAllDayInput').checked = true; // Reset 'All Day' checkbox
        document.getElementById('taskDueDate').value = ''; // Reset due date/time input
        document.getElementById('taskDueTime').value = ''; // Reset due date/time input
        form.classList.remove('was-validated'); // Remove validation class to reset the form state

        // Also hide the alert box
        const alertBox = document.getElementById('formErrorAlert');
        alertBox.classList.add('d-none');
        alertBox.textContent = ''; // Clear the error message
        const modalElement = new bootstrap.Modal(document.getElementById(this.modalId));
        // by default choose the Due On option

        // Apply default due date setting
        let selectElement = document.getElementById('dueDateTypeSelect');
        applyDefaultDueDate(selectElement);

        modalElement.show();
    }

    hide() {
        const modalElement = bootstrap.Modal.getInstance(document.getElementById(this.modalId));
        modalElement.hide();
    }

    showAlert(message, type) {
        const alertBox = document.getElementById('formErrorAlert');
        alertBox.textContent = message;
        alertBox.classList.remove('d-none'); // Show the alert box
        alertBox.classList.add(`alert-${type}`); // Use the 'type' to add specific styling, e.g., 'alert-danger' for errors
    }
}

function applyDefaultDueDate(selectElement) {
    // Change the value
    selectElement.value = 'due_on';
    let event = new Event('change', {
        'bubbles': true,    // Enable event bubbling
        'cancelable': true  // Event can be cancelled
    });
    // Dispatch the event
    selectElement.dispatchEvent(event);
}

class TaskDetailsModal {
    constructor(listManager, showAlert) {
        this.listManager = listManager;
        this.currentTask = null;
        this.modalId = 'taskDetailsModal';
        this.showAlertOutside = showAlert;
        this.modalElement = null;

        this.initModal();
    }

    initModal() {
        this.modalElement = new bootstrap.Modal(document.getElementById(this.modalId));
        this.setupEventListeners();
    }

    setListManager(listManager) {
        this.listManager = listManager;
    }

    setTaskUpdateCallback(callback) {
        this.onTaskUpdated = callback;
    }

    setupEventListeners() {
        const modalElement = document.getElementById(this.modalId);

        handleDueTypeChange(
            document.getElementById('dueDateTypeEditSelect'),
                'dateTimeEditOptions',
                'timeEditOptions',
                'taskEditDueDate',
                'taskEditAllDayInput'
        );

        document.getElementById('taskEditAllDayInput').addEventListener('change', function() {
            const timeOptions = document.getElementById('timeEditOptions');
            if (this.checked) {
                // hide the time input
                timeOptions.classList.add('d-none');
            } else {
                // show the time input
                timeOptions.classList.remove('d-none');
                // default to 9 AM
                document.getElementById('taskEditDueTime').value = '09:00';
            }
        });

        // Add an event listener for when the modal is fully hidden
        modalElement.addEventListener('hidden.bs.modal', () => {
            this.hide();
        });
        document.getElementById('updateTaskBtn').addEventListener('click', this.handleUpdateTask.bind(this));
    }

    setCurrentTask(task) {
        this.currentTask = task;
        this.populateModalFields(task);
    }

    populateModalFields(task) {
        document.getElementById('taskEditTitleInput').value = task.title;
        document.getElementById('taskList').textContent = task.list;
        document.getElementById('taskStatus').textContent = task.done ? 'Completed' : 'Pending';
        document.getElementById('taskCreated').textContent = formatDate(task.created);


        // Conditionally populate and display the "Due On" and "Due By" fields

        const due = task.due;
        const dueType = task.due_type;
        const allDay = task.all_day;

        document.getElementById('priorityEditSelect').value = task.priority;

        if (dueType === "due_on" || dueType === "due_by") {
            document.getElementById('dateTimeEditOptions').classList.remove('d-none');
            // prefill the due type select
            document.getElementById('dueDateTypeEditSelect').value = dueType;
            // prefill the date and time inputs
            let dateObj = new Date(due)
            // Obtain the browser's locale
            let browserLocale = navigator.language;
            // Format date for the date input, using the browser's locale
            let formattedDate = formatDateForm(dateObj.toISOString());
            // Format time for the time input, using the browser's locale
            let formattedTime = dateObj.toLocaleTimeString(browserLocale, {
                hour12: false, // Prefer 24-hour clock if supported by the locale
                hour: '2-digit',
                minute: '2-digit'
            });
            document.getElementById('taskEditDueDate').value = formattedDate;
            if (!allDay) {
                // uncheck the checkbox
                document.getElementById('taskEditAllDayInput').checked = false;
                // prefill and show the time input
                document.getElementById('timeEditOptions').classList.remove('d-none');
                document.getElementById('taskEditDueTime').value = formattedTime;
            } else {
                document.getElementById('timeEditOptions').classList.add('d-none');
                // check the checkbox
                document.getElementById('taskEditAllDayInput').checked = true;
            }
        } else {
            document.getElementById('dateTimeEditOptions').classList.add('d-none');
        }

        // Conditionally populate and display the "Done On" field
        const doneOn = task.done_on;

        const isTaskDone = task.done === 'true';
        document.getElementById('taskDoneOn').textContent = isTaskDone ? formatDate(doneOn) : '';
        document.getElementById('taskDoneOnLabel').style.display = isTaskDone ? 'block' : 'none';
        document.getElementById('taskDoneOn').style.display = isTaskDone ? 'block' : 'none';

    }

    async handleUpdateTask() {
        const titleInput = document.getElementById('taskEditTitleInput');
        const dueDateTypeSelect = document.getElementById('dueDateTypeEditSelect');
        const allDayCheckbox = document.getElementById('taskEditAllDayInput');
        const dueDateInput = document.getElementById('taskEditDueDate');
        const dueTimeInput = document.getElementById('taskEditDueTime');
        const prioritySelect = document.getElementById('priorityEditSelect');

        const title = titleInput.value.trim();
        const dueDateType = dueDateTypeSelect.value;
        const isAllDay = allDayCheckbox.checked;
        const dueDate = dueDateInput.value;
        const dueTime = dueTimeInput.value;
        // Format a date string in the format 'YYYY-MM-DDTHH:MM:SS'
        const dueDateTime = isAllDay ? `${dueDate}` : `${dueDate}T${dueTime}`;
        const priority = prioritySelect.value;

        let requestPayload = {
            title: title,
            priority: parseInt(priority,10),
        };

        // Adjust payload based on the due date type selection
        requestPayload.due_type = dueDateType;
        if (dueDateType !== 'none') {
            requestPayload.due = dueDateTime;
            requestPayload.all_day = isAllDay;
        } else {
            requestPayload.due = null;
            requestPayload.all_day = false;
        }

        try {
            await this.listManager.updateTask(this.currentTask.id, requestPayload);
            this.hide();
            this.onTaskUpdated(this.listManager.listByName(this.currentTask.list)); // Notify the parent UI manager of the change
            this.showAlertOutside('Task updated', 'success');
        } catch (error) {
            console.error('Error updating task:', error);
            this.showAlert(`Failed to update the task: ${error.message}`);
        }
    }


    show() {
        document.getElementById('formErrorAlert2').classList.add('d-none');
        this.modalElement.show();
    }

    hide() {
        this.modalElement.hide();
    }

    showAlert(message, type) {
        const alertBox = document.getElementById('formErrorAlert2');
        alertBox.textContent = message;
        alertBox.classList.remove('d-none'); // Show the alert box
        alertBox.classList.add(`alert-${type}`); // Use the 'type' to add specific styling, e.g., 'alert-danger' for errors
    }
}

function isOverdue(dueDate, currentDate, ignoreTime = false) {
    if (ignoreTime) {
        // Remove the time component from both dates
        const dueDateWithoutTime = new Date(dueDate.setHours(0, 0, 0, 0));
        const currentDateWithoutTime = new Date(currentDate.setHours(0, 0, 0, 0));

        // Compare only the dates (day-wise)
        return dueDateWithoutTime < currentDateWithoutTime;
    } else {
        // Compare including the time
        return new Date(dueDate) < new Date(currentDate);
    }
}

function handleDueTypeChange(dueDateTypeSelect, dateTimeOptionsId, timeOptionsId, dueDateInputId, allDayCheckboxId) {
    dueDateTypeSelect.addEventListener('change', function() {
        const dateTimeOptions = document.getElementById(dateTimeOptionsId);
        const timeOptions = document.getElementById(timeOptionsId);
        const dueDateInput = document.getElementById(dueDateInputId);
        const allDayCheckbox = document.getElementById(allDayCheckboxId);

        if (this.value === 'none') {
            dateTimeOptions.classList.add('d-none');
        } else {
            // Remove 'd-none' to show the options
            dateTimeOptions.classList.remove('d-none');
            // Check the 'All Day' checkbox by default
            allDayCheckbox.checked = true;
            dueDateInput.type = 'date'; // Default to date input
            timeOptions.classList.add('d-none'); // Hide the time input by default

            // Determine today's date
            const today = new Date();
            // Set the input value to today, adjusting format based on 'All Day'
            dueDateInput.value = today.toISOString().split('T')[0];
        }
    });
}

function applyListColour(list, element) {
    let opacityStart = list.filtered ? 0 : 0.5;
    let opacityEnd = 0;

// Define your gradient transition points relative to the viewport height
    let startVh = "0vh"; // Transition start point at 5% of the viewport height
    let endVh = "10vh"; // Transition end point at 10% of the viewport height

    element.style.backgroundImage = `linear-gradient(180deg, 
  rgba(${list.colour.r}, ${list.colour.g}, ${list.colour.b}, ${opacityStart}) 0%, 
  rgba(${list.colour.r}, ${list.colour.g}, ${list.colour.b}, ${opacityStart}) ${startVh}, 
  rgba(${list.colour.r}, ${list.colour.g}, ${list.colour.b}, ${opacityEnd}) ${endVh}, 
  rgba(${list.colour.r}, ${list.colour.g}, ${list.colour.b}, ${opacityEnd}) 100%)`;

}