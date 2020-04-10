compile:
	go run main.go
	dot -Tpng out.dot > out.png
	open out.png
