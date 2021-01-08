prepare_sampleapp:
	cp devapp/devapp.go	sampleapp/sampleapp.go
	cp devapp/devapp_test.go	sampleapp/sampleapp_test.go
	sed -i 's/package devapp/package sampleapp/' sampleapp/sampleapp.go
	sed -i 's/package devapp/package sampleapp/' sampleapp/sampleapp_test.go
	go run ./tsqb/ ./sampleapp/