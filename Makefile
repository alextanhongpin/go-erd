png:
	make compile-png

compile-%:
	# go run main.go | dot -T$*> out.$*
	go run main.go -in in.txt -out out.$*
	open out.$*

test:
	go run main.go
