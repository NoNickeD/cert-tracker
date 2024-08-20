import { Trend } from 'k6/metrics';

const apiResponseTime = new Trend('api_response_time');

export default function () {
    // Measure POST request
    const postRes = http.post('http://localhost:8080/check', postPayload, postParams);
    apiResponseTime.add(postRes.timings.duration);

    // Measure GET request for each domain
    domains.forEach((domain) => {
        const getRes = http.get(`http://localhost:8080/check?domain=${domain}`);
        apiResponseTime.add(getRes.timings.duration);
    });
}
