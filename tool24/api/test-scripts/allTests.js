import http from 'k6/http';
import { check, sleep } from 'k6';

// Define the base URL of your API
const BASE_URL = 'http://localhost:8080';

// Define test stages (optional, you can adjust the stages to your needs)
export const options = {
    stages: [
        { duration: '30s', target: 10 }, // ramp up to 10 users
        { duration: '1m', target: 10 }, // stay at 10 users for 3 minutes
        { duration: '30s', target: 0 },  // ramp down to 0 users
    ],
};

// Utility function to generate random IDs (adjust as needed)
function randomId() {
    return Math.floor(Math.random() * 1000).toString();
}

export default function () {
    // Test the GET /todos endpoint
    // let res = http.get(`${BASE_URL}/todos`);
    // check(res, {
    //     'GET /todos status 200': (r) => r.status === 200,
    // });
    // sleep(1);

    //Test the GET /todo/{id} endpoint
    let ids = [
      "7f9cd9f9-6b9e-46f1-a097-5956498ca85e",
      "2cebc463-bb73-4029-ba5d-5719ebfe42ea",
      "336231ee-82f5-4e29-9d88-4198821f635e",
      "5228fc5b-b453-48b3-bc08-c67fef025dd4",
      "761fb35a-a9c4-4dbf-827a-ac4676a17d8c",
      "ed7acaaa-8609-4092-8412-8ea6cbe52349",
      "d0830367-5025-477d-87ca-7fd804129a84",
      "d0fc3a3b-b0a5-4d47-81c3-eea29adb6d62",
      "e9a14c52-304f-497a-b5ce-3110fc425f39"
  ];
  
  ids.forEach(id => {
      let res = http.get(`${BASE_URL}/todo/${id}`);
  
      // Kiểm tra mã trạng thái HTTP
      check(res, {
          'GET /todo/{id} status 200 or 404': (r) => r.status === 200 || r.status === 404,
      });
  
      if (res.status === 200) {
          let jsonData = JSON.parse(res.body);
          check(jsonData, {
              'ID should match': (r) => r.id === id,
          });
      }
  
      sleep(1);
  });
  

    // Test the POST /todo endpoint

    // Test the PUT /todo/{id} endpoint
    // const updatePayload = JSON.stringify({
    //     title: `Updated Todo ${id}`,
    //     description: 'This is an updated test todo',
    //     status: 'completed',
    // });
    // res = http.put(`${BASE_URL}/todo/${id}`, updatePayload, {
    //     headers: { 'Content-Type': 'application/json' },
    // });
    // check(res, {
    //     'PUT /todo/{id} status 200 or 404': (r) => r.status === 200 || r.status === 404,
    // });
    // sleep(1);


    // Test the POST /todo/changeStatus/{id} endpoint
    // res = http.post(`${BASE_URL}/todo/changeStatus/${id}`);
    // check(res, {
    //     'POST /todo/changeStatus/{id} status 200 or 404': (r) => r.status === 200 || r.status === 404,
    // });
    // sleep(1);
}
