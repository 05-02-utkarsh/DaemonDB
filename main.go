package main

import (
	bplus "DaemonDB/bplustree"
	heapfile "DaemonDB/heapfile_manager"
	executor "DaemonDB/query_executor"
	codegen "DaemonDB/query_parser/code-generator"
	lex "DaemonDB/query_parser/lexer"
	"DaemonDB/query_parser/parser"
	"DaemonDB/wal_manager"
<<<<<<< HEAD

	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
=======
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
>>>>>>> 7e45a0c662a830cd99d00b75eb130fc9f506a3dc
)

type QueryRequest struct {
	Query string `json:"query"`
}

type QueryResponse struct {
	AST      string `json:"ast"`
	Bytecode string `json:"bytecode"`
	Result   string `json:"result"`
	Error    string `json:"error,omitempty"`
}

func main() {

	// ---- INIT DATABASE (same as before) ----

	walManager, err := wal_manager.OpenWAL("databases/demoDB/logs")
	if err != nil {
		log.Fatal(err)
	}

	pager := bplus.NewInMemoryPager()
	cache := bplus.NewBufferPool(10)
	tree := bplus.NewBPlusTree(pager, cache, bytes.Compare)

	heapFileManager, err := heapfile.NewHeapFileManager("databases/demoDB")
	if err != nil {
		walManager.Close()
		log.Fatal(err)
	}

	vm := executor.NewVM(tree, heapFileManager, walManager)
	defer vm.CloseIndexCache()

	if err := vm.RecoverAndReplayFromWAL(); err != nil {
		walManager.Close()
		log.Fatal(err)
	}

	// ---- HTTP HANDLER ----

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var req QueryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if line == "?" || strings.EqualFold(line, "help") {
			printHelp()
			continue
		}

<<<<<<< HEAD
		query := req.Query
=======
		query := line
>>>>>>> 7e45a0c662a830cd99d00b75eb130fc9f506a3dc

		// Lexer + Parser
		l := lex.New(query)
		p := parser.New(l)

		stmt, err := p.ParseStatement()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		// AST output
		astOutput := fmt.Sprintf("%#v", stmt)

<<<<<<< HEAD
		// Bytecode
		instructions := codegen.EmitBytecode(stmt)
		bytecodeOutput := ""
=======
		instructions, err := codegen.EmitBytecode(stmt)
		if err != nil {
			fmt.Printf("Codegen error: %v\n", err)
			continue
		}
>>>>>>> 7e45a0c662a830cd99d00b75eb130fc9f506a3dc
		for i, instr := range instructions {
			bytecodeOutput += fmt.Sprintf(
				"%d: OP=%v, VALUE=%v\n",
				i, instr.Op, instr.Value,
			)
		}

<<<<<<< HEAD
		// Execute
		err = vm.Execute(instructions)

		response := QueryResponse{
			AST:      astOutput,
			Bytecode: bytecodeOutput,
=======
		fmt.Println("\n=== Execution ===")
		if err := vm.Execute(instructions); err != nil {
			fmt.Printf("Execution error: %v\n", err)
>>>>>>> 7e45a0c662a830cd99d00b75eb130fc9f506a3dc
		}

		if err != nil {
			response.Error = err.Error()
		} else {
			response.Result = "Execution successful"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Println("DaemonDB server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func printHelp() {
	fmt.Println("Supported commands:")
	fmt.Println("  SHOW DATABASES")
	fmt.Println("  CREATE DATABASE <name>")
	fmt.Println("  USE <database>")
	fmt.Println("  CREATE TABLE <name> ( col type [primary key], ... )")
	fmt.Println("  INSERT INTO <table> VALUES ( val1, val2, ... )")
	fmt.Println("  SELECT * FROM <table> [ WHERE col = value ]")
	fmt.Println("  SELECT * FROM t1 [ INNER|LEFT|RIGHT|FULL ] JOIN t2 ON col1 = col2 [ WHERE ... ]")
	fmt.Println("  BEGIN; COMMIT; ROLLBACK")
	fmt.Println("  exit")
	fmt.Println("Note: UPDATE/DELETE/DROP are parsed but not executed yet.")
}
