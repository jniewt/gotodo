export function sortTasks(items, sortStrategy) {
    return items.sort(sortStrategy);
}

export function sortByTitleThenDone(a, b) {
    // Move done items to the bottom
    if (a.done !== b.done) {
        return a.done ? 1 : -1;
    }
    // Then sort by title
    return a.title.localeCompare(b.title);
}