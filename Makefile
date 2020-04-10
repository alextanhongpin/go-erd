png:
	make compile-png

compile-%:
	go run main.go | dot -T$*> out.$*
	open out.$*

test:
	go run main.go
