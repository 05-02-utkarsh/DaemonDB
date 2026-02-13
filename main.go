package main

import (
	bplus "DaemonDB/bplustree"
	heapfile "DaemonDB/heapfile_manager"
	executor "DaemonDB/query_executor"
	codegen "DaemonDB/query_parser/code-generator"
	lex "DaemonDB/query_parser/lexer"
	"DaemonDB/query_parser/parser"
	"DaemonDB/wal_manager"

	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

		query := req.Query

		// Lexer + Parser
		l := lex.New(query)
		p := parser.New(l)

		stmt := p.ParseStatement()

		// AST output
		astOutput := fmt.Sprintf("%#v", stmt)

		// Bytecode
		instructions := codegen.EmitBytecode(stmt)
		bytecodeOutput := ""
		for i, instr := range instructions {
			bytecodeOutput += fmt.Sprintf(
				"%d: OP=%v, VALUE=%v\n",
				i, instr.Op, instr.Value,
			)
		}

		// Execute
		err = vm.Execute(instructions)

		response := QueryResponse{
			AST:      astOutput,
			Bytecode: bytecodeOutput,
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
