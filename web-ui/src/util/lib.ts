/**
 * Returns a random number between min (inclusive) and max (exclusive)
 */
export const getRandomArbitrary = (min: number, max: number): number => {
    return Math.random() * (max - min) + min;
}

/**
 * Returns a random integer between min (inclusive) and max (inclusive).
 * The value is no lower than min (or the next integer greater than min
 * if min isn't an integer) and no greater than max (or the next integer
 * lower than max if max isn't an integer).
 * Using Math.round() will give you a non-uniform distribution!
 */
export const getRandomInt = (min: number, max: number): number => {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

export const isInteger = (str?: string): boolean => {
    return Number.isInteger(str ? +str : undefined);
}

export const formatValue = (megabytes: number, shortUnitForm = false) => {
    let val: number;
    let unitSuffix: string;

    if (megabytes < 1024) {
        val = megabytes;

        if (shortUnitForm) {
            unitSuffix = "MB";
        } else {
            unitSuffix = val === 1 ? 'Megabyte' : 'Megabytes';
        }

        return `${megabytes} ${unitSuffix}`;
    } else if (megabytes < 1024 * 1024) {
        val = parseFloat((megabytes / 1024).toFixed(2));

        if (shortUnitForm) {
            unitSuffix = "GB";
        } else {
            unitSuffix = val === 1 ? 'Gigabyte' : 'Gigabytes';
        }

        return `${(megabytes / 1024).toFixed(2)} ${unitSuffix}`;
    } else {
        val = parseFloat((megabytes / (1024 * 1024)).toFixed(2));

        if (shortUnitForm) {
            unitSuffix = "TB";
        } else {
            unitSuffix = val === 1 ? 'Terabyte' : 'Terabytes';
        }

        return `${(megabytes / (1024 * 1024)).toFixed(2)} ${unitSuffix}`;
    }
};
