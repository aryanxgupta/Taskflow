echo "--- Starting TaskFlow Concurrency Test ---"
echo "Sending 6 requests at the same time..."
echo ""

curl -X POST http://localhost:8080/tasks -d '{"payload": "https://httpbin.org/delay/1"}' &


curl -X POST http://localhost:8080/tasks -d '{"payload": "https://httpbin.org/delay/7"}' &

curl -X POST http://localhost:8080/tasks -d '{"payload": "https://httpbin.org/status/404"}' &

curl -X POST http://localhost:8080/tasks -d '{"payload": "http://this-domain-does-not-exist-12345.com"}' &

curl -X POST http://localhost:8080/tasks -d '{"payload": 12345}' &

curl -X POST http://localhost:8080/tasks -d '{"payload": "this is bad json' &

echo "All 6 requests sent. Waiting for all jobs to complete..."
wait
echo ""
echo "--- Test Complete ---"
echo "Check your server logs and use GET requests to see the results."