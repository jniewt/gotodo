async function fetchAllLists() {
    try {
        const response = await fetch('/api/list');
        const data = await response.json();
        const listsDisplay = document.getElementById('listsDisplay');
        listsDisplay.innerHTML = ''; // Clear current lists display

        data.lists.forEach(listName => {
            const listElement = document.createElement('a');
            listElement.classList.add('list-group-item', 'list-group-item-action', 'd-flex', 'justify-content-between', 'align-items-center');
            listElement.textContent = listName;
            listElement.setAttribute('href', '#');
            listElement.onclick = function() {
                fetchListDetails(listName);
                return false; // Prevent default anchor action
            };

            // Dropdown for the 3-dot menu
            const dropdown = document.createElement('div');
            dropdown.classList.add('dropdown');
            dropdown.innerHTML = `
        <a class="text-secondary" href="#" role="button" data-bs-toggle="dropdown" aria-expanded="false">
        <i class="bi bi-three-dots-vertical"></i>
        </a>
        <ul class="dropdown-menu" aria-labelledby="dropdownMenuButton">
            <li><a class="dropdown-item" href="#" onclick="event.stopPropagation(); deleteList('${listName}');">Delete</a></li>
        </ul>
    `;
            // Stop propagation for the dropdown to prevent list selection
            dropdown.addEventListener('click', function(event) {
                event.stopPropagation();
            });

            listElement.appendChild(dropdown);
            listsDisplay.appendChild(listElement);
        });
    } catch (error) {
        console.error('Failed to fetch lists:', error);
    }
}

fetchAllLists(); // Call the function to fetch lists when the page loads

function displayListDetails(list) {
    const listDetailsEl = document.getElementById('listItems');
    listDetailsEl.setAttribute('data-current-list', list.name); // Store the current list name in the container
    listDetailsEl.style.position = 'relative';
    listDetailsEl.innerHTML = `<h3>${list.name}</h3>`; // Display the list name

    if (!list.items || list.items.length === 0) {
        listDetailsEl.innerHTML += '<p>No items in this list</p>';
    } else {
        const sortedItems = sortItems(list.items, sortByTitleThenDone); // Sort items
        const itemsEl = document.createElement('ul');
        itemsEl.classList.add('list-group');

        sortedItems.forEach(item => {
            const itemEl = document.createElement('li');
            itemEl.classList.add('list-group-item', 'd-flex', 'align-items-center');
            itemEl.setAttribute('data-item-id', item.id);
            itemEl.setAttribute('data-title', item.title);
            itemEl.setAttribute('data-list', list.name);
            itemEl.setAttribute('data-done', item.done);
            itemEl.setAttribute('data-created', item.created);
            itemEl.setAttribute('data-all-day', item.all_day);
            itemEl.setAttribute('data-due-on', item.due_on || '');
            itemEl.setAttribute('data-due-by', item.due_by || '');
            itemEl.setAttribute('data-done-on', item.done_on || '')
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.classList.add('me-2');
            checkbox.checked = item.done;

            const titleSpan = document.createElement('span');
            titleSpan.textContent = item.title;
            if (item.done) {
                titleSpan.classList.add('text-decoration-line-through', 'text-muted'); // Bootstrap classes for strikethrough and color fade
            }

            checkbox.onchange = function() {
                if (checkbox.checked) {
                    titleSpan.classList.add('text-decoration-line-through', 'text-muted');
                } else {
                    titleSpan.classList.remove('text-decoration-line-through', 'text-muted');
                }

                // handle the sorting, animation and model update
                handleCheckboxChange(list, item.id, checkbox.checked, itemEl);
            };

            itemEl.appendChild(checkbox);
            itemEl.appendChild(titleSpan);
            itemsEl.appendChild(itemEl);
        });

        listDetailsEl.appendChild(itemsEl);
    }

    // Create and append the add task button
    const addButton = document.createElement('button');
    addButton.innerHTML = '<i class="bi bi-plus"></i>'; // Using Bootstrap Icons for the plus icon
    addButton.classList.add('btn', 'btn-primary', 'add-item-btn', 'rounded-circle');
    addButton.setAttribute('data-bs-toggle', 'modal');
    addButton.setAttribute('data-bs-target', '#addItemModal'); // Assuming your modal for adding items is ready

    document.body.appendChild(addButton); // Append to body to ensure fixed positioning relative to viewport

}

async function createList(listName) {
    try {
        const response = await fetch('/api/list', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name: listName }),
        });
        if (response.ok) {
            fetchAllLists(); // Refresh the lists to include the new list
            showAlert('List created successfully', 'success')
        } else {
            console.error('Failed to create list');
            const errorResponse = await response.json(); // Assuming the server responds with JSON
            const errorMessage = errorResponse.error || 'An unexpected error occurred'; // Fallback error message
            showAlert(`Failed to create list: ${errorMessage}`); // Show the error from the server

        }
    } catch (error) {
        console.error('Failed to create list:', error);
        showAlert(`Failed to create list: ${error.message}`, 'danger'); // Show the error from the catch block
    }
}

async function deleteList(listName) {
    try {
        // Fetch list details first to check if it has items
        const responseDetails = await fetch(`/api/list/${listName}`);
        const listDetails = await responseDetails.json();

        // Check if list is not empty
        if (listDetails.list.items && listDetails.list.items.length > 0) {
            // Ask for confirmation
            const isConfirmed = confirm(`The list "${listName}" is not empty. Are you sure you want to delete it?`);
            if (!isConfirmed) {
                return; // Stop if user does not confirm
            }
        }

        // Proceed with deletion if list is empty or user confirmed
        const responseDelete = await fetch(`/api/list/${listName}`, {
            method: 'DELETE',
        });
        if (responseDelete.ok) {
            alert('List deleted successfully.');
            fetchAllLists(); // Refresh the lists display
            showAlert('List deleted successfully', 'success')
        } else {
            console.error('Failed to delete list');
            const errorResponse = await responseDelete.json(); // Assuming the server responds with JSON
            const errorMessage = errorResponse.error || 'An unexpected error occurred'; // Fallback error message
            showAlert(`Failed to delete list: ${errorMessage}`); // Show the error from the server
        }
    } catch (error) {
        console.error('Failed to delete list:', error);
        showAlert(`Failed to delete list: ${error.message}`, 'danger'); // Show the error from the catch block
    }
}

async function deleteTask(taskId) {
    try {
        const response = await fetch(`/api/items/${taskId}`, { method: 'DELETE' });
        if (response.ok) {
            // Remove the task from the list or refresh the list
            console.log('Task deleted successfully');
            const listDetailsEl = document.getElementById('listItems');
            const listName = listDetailsEl.getAttribute('data-current-list');
            fetchListDetails(listName); // Assuming this function refreshes the task list
            showAlert('Task deleted successfully', 'success')
        } else {
            console.error('Failed to delete task');
            const errorResponse = await response.json(); // Assuming the server responds with JSON
            const errorMessage = errorResponse.error || 'An unexpected error occurred'; // Fallback error message
            showAlert(`Failed to delete task: ${errorMessage}`); // Show the error from the server
        }
    } catch (error) {
        console.error('Error deleting task:', error);
        showAlert(`Failed to delete the task: ${error.message}`); // Show the error from the catch block
    }
}

async function fetchListDetails(listName) {
    console.log(`Fetching details for list: ${listName}`);
    try {
        const response = await fetch(`/api/list/${listName}`);
        if (!response.ok) {
            throw new Error('Failed to fetch list details');
        }
        const listDetails = await response.json();
        console.log('List details fetched:', listDetails);

        displayListDetails(listDetails.list); // Assuming you have a function to display these details
    } catch (error) {
        console.error('Error fetching list details:', error);
    }
}

async function populateListDropdown() {
    try {
        const response = await fetch('/api/list');
        const data = await response.json();
        const dropdown = document.getElementById('taskListDropdown');
        dropdown.innerHTML = ''; // Clear existing options
        data.lists.forEach(listName => {
            const option = document.createElement('option');
            option.value = listName;
            option.textContent = listName;
            dropdown.appendChild(option);
        });
    } catch (error) {
        console.error('Failed to fetch lists:', error);
    }
}

function handleCheckboxChange(list, itemId, isChecked, itemElement) {
    // Apply fade-out effect for visual feedback
    itemElement.classList.add('fade-out');

    itemElement.addEventListener('animationend', async () => {
        // PATCH the new 'done' status to the server
        try {
            const response = await fetch(`/api/items/${itemId}`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ done: isChecked }),
            });

            if (!response.ok) {
                throw new Error('Failed to update item status');
            }

            const updatedTask = await response.json(); // Assuming the server returns the updated task

            const itemIndex = list.items.findIndex(item => item.id === itemId);
            if (itemIndex !== -1) {
                // Update the local model with the new data from the server
                list.items[itemIndex] = updatedTask.task;

                // Also update the item element's data attributes to reflect the new state
                itemElement.setAttribute('data-done', updatedTask.task.done);
                itemElement.setAttribute('data-done-on', updatedTask.task.done_on || '');
            }

            // Re-fetch and display the updated list to ensure UI consistency
            displayListDetails(list);

            // Remove fade-out class to reset element's state for future animations
            itemElement.classList.remove('fade-out');

        } catch (error) {
            console.error('Error updating item status:', error);
            // Optionally, handle the error (e.g., display a message to the user)
        }
    }, { once: true });
}

document.getElementById('newListForm').addEventListener('submit', async function(event) {
    event.preventDefault();
    const listName = document.getElementById('listNameInput').value;
    // Your createList function logic here
    if (listName) {
        await createList(listName);
        fetchAllLists(); // Refresh the lists display
        bootstrap.Modal.getInstance(document.getElementById('newListModal')).hide(); // Hide the modal
    }
});

// focus on the input field when the modal is shown
document.getElementById('newListModal').addEventListener('shown.bs.modal', function () {
    document.getElementById('listNameInput').focus();
});

// clear the form when the modal is hidden
document.getElementById('newListModal').addEventListener('hidden.bs.modal', function () {
    document.getElementById('newListForm').reset();
});

// populate the dropdown when the DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    populateListDropdown();
});

document.getElementById('addItemModal').addEventListener('show.bs.modal', function (event) {
    populateListDropdown();
    // Get the title input field
    const titleInput = document.getElementById('taskTitleInput');

    // Clear and focus the title input field
    titleInput.value = '';
    titleInput.focus();
});

document.getElementById('addItemModal').addEventListener('hidden.bs.modal', function() {
    // Reset the due date type to 'None'
    document.getElementById('dueDateTypeSelect').value = 'none';

    // Hide the date/time options section
    document.getElementById('dateTimeOptions').classList.add('d-none');

    // Reset and check the 'All Day' checkbox
    const allDayCheckbox = document.getElementById('taskAllDayInput');
    allDayCheckbox.checked = true;
    allDayCheckbox.dispatchEvent(new Event('change')); // Trigger change event to adjust the UI accordingly

    // Clear the date/time input
    document.getElementById('taskDueDateTime').value = '';

    // Reset the title input
    document.getElementById('taskTitleInput').value = '';

});

document.getElementById('deleteTask').addEventListener('click', function() {
    // Use taskContextMenu's dataset since the deleteTask li doesn't have its own dataset
    const taskId = document.getElementById('taskContextMenu').dataset.taskId;
    if (taskId) {
        deleteTask(taskId); // Implement this function to delete the task
    }
});

document.getElementById('saveTaskButton').addEventListener('click', async () => {
    const titleInput = document.getElementById('taskTitleInput');
    const listDropdown = document.getElementById('taskListDropdown');
    const dueDateTypeSelect = document.getElementById('dueDateTypeSelect');
    const allDayCheckbox = document.getElementById('taskAllDayInput');
    const dueDateTimeInput = document.getElementById('taskDueDateTime');

    const title = titleInput.value;
    const listName = listDropdown.value;
    const dueDateType = dueDateTypeSelect.value;
    const isAllDay = allDayCheckbox.checked;
    const dueDateTime = dueDateTimeInput.value;

    // Initialize the request payload with title and all_day
    let requestPayload = {
        title: title,
    };

    // Conditionally add dueOn or dueBy based on the dropdown selection
    if (dueDateType === 'dueOn') {
        requestPayload.due_on = dueDateTime;
        requestPayload.all_day = isAllDay;
    } else if (dueDateType === 'dueBy') {
        requestPayload.due_by = dueDateTime;
        requestPayload.all_day = isAllDay;
    }

    try {
        const response = await fetch(`/api/list/${listName}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestPayload),
        });

        const data = await response.json();

        if (response.ok) {
            console.log('Task added:', data.task);
            fetchListDetails(listName); // Refresh the task list
            bootstrap.Modal.getInstance(document.getElementById('addItemModal')).hide(); // Hide the modal
            showAlert('Task added successfully', 'success');
        } else {
            console.error('Failed to add task');
            showAlert(`Failed to add the task: ${data.error}`); // Show error alert
        }
    } catch (error) {
        console.error('Error adding task:', error);
        showAlert(`Failed to add the task: ${error.message}`);
    }
});

// Add a context menu to the task list items
document.addEventListener('DOMContentLoaded', () => {
    // Assuming the taskList element itself is not dynamically added/removed
    const taskList = document.getElementById('listItems');

    taskList.addEventListener('contextmenu', function(event) {
        // Check if the clicked element is a task entry
        const clickedTask = event.target.closest('.list-group-item');
        if (!clickedTask) {
            // If the click did not happen on a task entry, ignore it
            return;
        }

        event.preventDefault();

        const taskContextMenu = document.getElementById('taskContextMenu');
        // Set the position of the menu
        taskContextMenu.style.left = `${event.pageX}px`;
        taskContextMenu.style.top = `${event.pageY}px`;
        taskContextMenu.style.display = 'block';

        // Store the task ID, if needed for deletion
        taskContextMenu.dataset.taskId = clickedTask.getAttribute('data-item-id');

        // Hide the menu when clicking elsewhere
        document.addEventListener('click', () => taskContextMenu.style.display = 'none', { once: true });
    });
});

// Handle task details modal
document.getElementById('listItems').addEventListener('click', function(event) {
    if (event.target.type === 'checkbox') {
        return; // Ignore clicks on checkboxes
    }
    const taskItem = event.target.closest('.list-group-item');
    if (taskItem) {
        // Assuming you store task details as data attributes or retrieve them here
        const taskId = taskItem.getAttribute('data-id');
        // Fetch or otherwise retrieve the full task details using taskId if not already stored in data attributes

        // Here we're assuming task details are directly available
        document.getElementById('taskTitle').textContent = taskItem.getAttribute('data-title');
        document.getElementById('taskList').textContent = taskItem.getAttribute('data-list');
        document.getElementById('taskStatus').textContent = taskItem.getAttribute('data-done') === 'true' ? 'Done' : 'Not Done';
        document.getElementById('taskCreated').textContent = formatDate(taskItem.getAttribute('data-created'));
        document.getElementById('taskAllDay').textContent = taskItem.getAttribute('data-all-day') === 'true' ? 'Yes' : 'No';

        // Conditionally populate and display the "Due On" and "Due By" fields
        const allDay = taskItem.getAttribute('data-all-day') === 'true';
        const dueOn = taskItem.getAttribute('data-due-on');
        const dueBy = taskItem.getAttribute('data-due-by');

        document.getElementById('taskDueOn').textContent = dueOn ? formatDateHumanReadable(dueOn, allDay) : '';
        document.getElementById('taskDueOn').style.display = dueOn ? 'block' : 'none';
        document.getElementById('taskDueOnLabel').style.display = dueOn ? 'block' : 'none';

        document.getElementById('taskDueBy').textContent = dueBy ? formatDateHumanReadable(dueBy, allDay) : '';
        document.getElementById('taskDueBy').style.display = dueBy ? 'block' : 'none';
        document.getElementById('taskDueByLabel').style.display = dueBy ? 'block' : 'none';

        // Conditionally populate and display the "Done On" field
        const doneOn = taskItem.getAttribute('data-done-on');
        const isTaskDone = taskItem.getAttribute('data-done') === 'true';
        document.getElementById('taskDoneOn').textContent = isTaskDone ? formatDate(doneOn) : '';
        document.getElementById('taskDoneOnLabel').style.display = isTaskDone ? 'block' : 'none';
        document.getElementById('taskDoneOn').style.display = isTaskDone ? 'block' : 'none';

        // Show the modal
        var taskDetailsModal = new bootstrap.Modal(document.getElementById('taskDetailsModal'));
        taskDetailsModal.show();
    }
});

document.getElementById('dueDateTypeSelect').addEventListener('change', function() {
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
            dueDateTimeInput.value = `${currentValue}T00:00`; // Default to midnight
        }
        // Note: If there was already a datetime value, changing the input type back to 'datetime-local'
        // will naturally preserve it, so there's no need to explicitly set it again.
    }
});



function sortItems(items, sortStrategy) {
    return items.sort(sortStrategy);
}

function sortByTitleThenDone(a, b) {
    // Move done items to the bottom
    if (a.done !== b.done) {
        return a.done ? 1 : -1;
    }
    // Then sort by title
    return a.title.localeCompare(b.title);
}

function showAlert(message, type = 'danger') {
    const alertPlaceholder = document.getElementById('alertPlaceholder');

    // Optional: Clear existing alerts to avoid over-crowding
    alertPlaceholder.innerHTML = ''; // This line clears previous alerts. Remove if you prefer to stack alerts.

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

function formatDate(isoDateString) {
    if (!isoDateString) return 'N/A'; // Handle null or undefined dates
    const date = new Date(isoDateString);
    return date.toLocaleString(); // Adjust formatting as needed
}

function formatDateHumanReadable(isoDateString, allDay = false) {
    if (!isoDateString) return 'N/A'; // Handle null or undefined dates

    const date = new Date(isoDateString);
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);

    // Check if the date is in the current year
    const isCurrentYear = now.getFullYear() === date.getFullYear();

    // Setting up date options based on whether it's all day and if it's the current year
    const dateOptions = {
        month: 'long',
        day: 'numeric',
        ...(isCurrentYear ? {} : { year: 'numeric' }), // Add year if it's not the current year
    };

    // Setting up time options based on whether it's all day
    const timeOptions = allDay ? {} : {
        hour: '2-digit',
        minute: '2-digit',
    };

    // Formatting date without the time if it's all day
    if (allDay) {
        const formattedDate = date.toLocaleDateString('default', dateOptions);
        if (date.toDateString() === today.toDateString()) {
            return 'Today';
        } else if (date.toDateString() === yesterday.toDateString()) {
            return 'Yesterday';
        } else if (date.toDateString() === tomorrow.toDateString()) {
            return 'Tomorrow';
        } else {
            return formattedDate;
        }
    } else {
        // Including time in the formatted date
        const formattedDateTime = date.toLocaleString('default', {...dateOptions, ...timeOptions});
        if (date.toDateString() === today.toDateString()) {
            return `Today, ${date.toLocaleTimeString('default', timeOptions)}`;
        } else if (date.toDateString() === yesterday.toDateString()) {
            return `Yesterday, ${date.toLocaleTimeString('default', timeOptions)}`;
        } else if (date.toDateString() === tomorrow.toDateString()) {
            return `Tomorrow, ${date.toLocaleTimeString('default', timeOptions)}`;
        } else {
            return formattedDateTime;
        }
    }
}

