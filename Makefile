serve-client:
	cd client; npx http-server -c-1 . --ssl --key ./localhost-key.pem --cert ./localhost.pem

serve-tracker:
	cd tracker; go run main.go