// Test scenarios configurations
export const testScenarios = {
    smoke: {
        executor: "constant-vus",
        vus: 1,
        duration: "1m",
    },
    load: {
        executor: "ramping-vus",
        startVUs: 10,
        stages: [
            { duration: "5m", target: 50 },
            { duration: "10m", target: 50 },
            { duration: "5m", target: 10 },
        ],
    },
    stress: {
        executor: "ramping-vus",
        startVUs: 10,
        stages: [
            { duration: "2m", target: 100 },
            { duration: "5m", target: 100 },
            { duration: "2m", target: 200 },
            { duration: "5m", target: 200 },
            { duration: "2m", target: 300 },
            { duration: "5m", target: 0 },
        ],
    },
    spike: {
        executor: "ramping-vus",
        startVUs: 10,
        stages: [
            { duration: "1m", target: 10 },
            { duration: "1m", target: 100 },
            { duration: "3m", target: 100 },
            { duration: "1m", target: 10 },
        ],
    },
    soak: {
        executor: "constant-vus",
        vus: 50,
        duration: "6h",
    },
};