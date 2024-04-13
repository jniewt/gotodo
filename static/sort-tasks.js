export function sortTasks(items, sortStrategies) {
    return items.sort((a, b) => {
        for (let strategy of sortStrategies) {
            const result = strategy(a, b);
            if (result !== 0) return result;
        }
        return 0;
    });
}

export function sortByDone(a, b) {
    return a.done === b.done ? 0 : a.done ? 1 : -1;
}

export function sortByPriority(a, b) {
    return b.priority - a.priority;
}

export function sortByDueDate(a, b) {
    let a_due = a.due;
    let b_due = b.due;
    if (a_due === b_due) return 0; // Handles equal due dates, including no due date
    if (!a_due) return 1; // No due date goes to the bottom
    if (!b_due) return -1; // No due date goes to the bottom

    // compare just days
    let a_date = new Date(a_due);
    let b_date = new Date(b_due);
    let dayCmp = a_date.setHours(0, 0, 0, 0) - b_date.setHours(0, 0, 0, 0);
    if (dayCmp !== 0) return dayCmp;

    // same day, so check if any task is all day
    let a_allDay = a.all_day || false;
    let b_allDay = b.all_day || false;

    if (a_allDay && b_allDay) return 0; // both all day
    // if a is all day, it goes after b
    if (a_allDay) return 1;
    // if b is all day, it goes after a
    if (b_allDay) return -1;

    // compare times
    return new Date(a_due) - new Date(b_due); // Earlier due dates come first
}

export function sortByTitle(a, b) {
    return a.title.localeCompare(b.title);
}