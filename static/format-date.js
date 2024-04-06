export function formatDate(isoDate) {
    if (!isoDate) return 'N/A';
    const date = new Date(isoDate);
    return date.toLocaleString();
}

// Takes an ISO date string and returns a date string in the format 'YYYY-MM-DD' that is expected by the date input element
export function formatDateForm(isoDate) {
    // Your original Date object
    let dateObj = new Date(isoDate);

// Get local date components
    let year = dateObj.getFullYear();
    let month = dateObj.getMonth() + 1; // getMonth() returns 0-11
    let day = dateObj.getDate();

// Ensure month and day are in two-digit format
    month = month < 10 ? '0' + month : month;
    day = day < 10 ? '0' + day : day;

// Construct the formatted date string in 'YYYY-MM-DD' format
    return `${year}-${month}-${day}`;
}

export function formatDateHuman(isoDate, ignoreTime = false) {
    if (!isoDate) return 'N/A'; // Handle null or undefined dates

    const date = new Date(isoDate);
    const today = new Date();
    today.setHours(0, 0, 0, 0); // Normalize today to start of day

    const tomorrow = new Date();
    tomorrow.setHours(0, 0, 0, 0);
    tomorrow.setDate(tomorrow.getDate() + 1);

    const yesterday = new Date();
    yesterday.setHours(0, 0, 0, 0);
    yesterday.setDate(yesterday.getDate() - 1);

    // Check if the date is in the current year
    const isCurrentYear = (new Date).getFullYear() === date.getFullYear();

    // Setting up date options based on whether it's all day and if it's the current year
    const dateOptions = {
        month: 'long',
        day: 'numeric',
        ...(isCurrentYear ? {} : { year: 'numeric' }), // Add year if it's not the current year
    };

    // Setting up time options based on whether it's all day
    const timeOptions = ignoreTime ? {} : {
        hour: '2-digit',
        minute: '2-digit',
    };

    if (ignoreTime) { // Formatting date without the time if ignoreTime is true
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
        if (date.toDateString() === today.toDateString()) { // If the date is today and time is not ignored, return the time only
            return `${date.toLocaleTimeString('default', timeOptions)}`;
        } else if (date.toDateString() === yesterday.toDateString()) { // If the date is yesterday
            return `Yesterday, ${date.toLocaleTimeString('default', timeOptions)}`;
        } else if (date.toDateString() === tomorrow.toDateString()) { // If the date is tomorrow
            return `Tomorrow, ${date.toLocaleTimeString('default', timeOptions)}`;
        } else { // Otherwise return full date and time
            return formattedDateTime;
        }
    }
}
