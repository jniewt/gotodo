export function formatDate(isoDateString) {
    if (!isoDateString) return 'N/A'; // Handle null or undefined dates
    const date = new Date(isoDateString);
    return date.toLocaleString(); // Adjust formatting as needed
}

export function formatDateHumanReadable(isoDateString, allDay = false) {
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
