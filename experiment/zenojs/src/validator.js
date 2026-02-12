
// ZenoJS Validator

const RULES = {
    required: (val) => val !== null && val !== undefined && val !== '',
    email: (val) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val),
    numeric: (val) => !isNaN(parseFloat(val)) && isFinite(val),
    min: (val, arg) => (typeof val === 'string' ? val.length : val) >= parseFloat(arg),
    max: (val, arg) => (typeof val === 'string' ? val.length : val) <= parseFloat(arg)
};

const MESSAGES = {
    required: 'This field is required.',
    email: 'Please enter a valid email.',
    numeric: 'This field must be a number.',
    min: 'Minimum value is :arg.',
    max: 'Maximum value is :arg.'
};

export function validate(data, rules) {
    const errors = {};
    let isValid = true;

    for (const field in rules) {
        const fieldRules = rules[field].split('|'); // "required|min:3"
        const value = getNestedValue(data, field);

        for (const ruleStr of fieldRules) {
            const [ruleName, ruleArg] = ruleStr.split(':');
            const ruleFn = RULES[ruleName];

            if (ruleFn) {
                // Skip validation if optional and empty (except required)
                if (ruleName !== 'required' && (value === null || value === undefined || value === '')) {
                    continue;
                }

                if (!ruleFn(value, ruleArg)) {
                    const msg = MESSAGES[ruleName] || 'Invalid field.';
                    errors[field] = msg.replace(':arg', ruleArg);
                    isValid = false;
                    break; // One error per field
                }
            } else {
                console.warn(`[Validator] Unknown rule: ${ruleName}`);
            }
        }
    }

    return { isValid, errors };
}

function getNestedValue(obj, path) {
    return path.split('.').reduce((o, key) => (o && o[key] !== undefined) ? o[key] : undefined, obj);
}
