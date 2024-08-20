import http from "k6/http";
import { check, sleep } from "k6";
import { testScenarios } from "./config.js";

export let options = {
  scenarios: {
    currentTest: testScenarios["smoke"], // Choose the test type here
  },
};

const domains = [
  "vodafone.gr",
  "vodafone.com",
  "vodafonecu.gr",
  "youtube.com",
  "google.com",
  "srekubecraft.io",
  "example.com",
];

export default function () {
  let allChecksPassed = true; // Track overall success

  domains.forEach((domain) => {
    const getRes = http.get(`http://localhost:8080/check?domain=${domain}`);

    // Log the response for visibility
    console.log(`Response for domain ${domain}: ${getRes.body}`);

    // Parse the response JSON
    const getResult = getRes.json();

    // Validate the structure of the response
    const statusCheck = getRes.status === 200;
    if (!statusCheck) {
      console.error(
        `Status check failed for domain: ${domain}, received status: ${getRes.status}`
      );
    }

    // Validate that required fields are present and of the correct type
    const responseCheck =
      typeof getResult.domain === "string" &&
      typeof getResult.expiry_date === "string" &&
      typeof getResult.issued_date === "string" &&
      typeof getResult.days_remaining === "number" &&
      typeof getResult.certificate_authority === "string" &&
      getResult.http_status === 200;
    if (!responseCheck) {
      console.error(
        `Response check failed for domain: ${domain}. Response data: ${JSON.stringify(
          getResult
        )}`
      );
    }

    // Aggregate the overall check results
    allChecksPassed = allChecksPassed && statusCheck && responseCheck;

    // Pause between each GET request to avoid overwhelming the server
    sleep(1);
  });

  // Validate the final result
  check(null, {
    "All GET /check requests returned status 200 and responses are valid": () =>
      allChecksPassed,
  });
}
