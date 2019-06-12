package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Block contains data that will be written to the blockchain.
type Block struct {
	Position  int
	Data      Transaction
	Timestamp string
	Hash      string
	PrevHash  string
}

// Transaction contains data for a checked out MedicalRecord
type Transaction struct {
	WalletAddress string `json:"wallet_address"`
	UserID        string `json:"user_id"`
	UserRole      string `json:"user_role"`
	UpdatedKey    string `json:"updated_key"`
	UpdatedValue  string `json:"updated_value"`
	IsGenesis     bool   `json:"is_genesis"`
}

// User contains data about user and role
type User struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// MedicalRecord is the Wallet, containing data about patient's medical record
type MedicalRecord struct {
	WalletAddress string   `json:"wallet_address"`
	FullName      string   `json:"full_name"`
	Operations    []string `json:"operations"`
	Prescriptions []string `json:"prescriptions"`
	Allergies     []string `json:"allergies"`
	CreationDate  string   `json:"creation_date"`
}

func (b *Block) generateHash() {
	// get string val of the Data
	bytes, _ := json.Marshal(b.Data)
	// concatenate the dataset
	data := string(b.Position) + b.Timestamp + string(bytes) + b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func CreateBlock(prevBlock *Block, transaction Transaction) *Block {
	block := &Block{}
	block.Position = prevBlock.Position + 1
	block.Timestamp = time.Now().String()
	block.Data = transaction
	block.PrevHash = prevBlock.Hash
	block.generateHash()

	return block
}

// Blockchain is an ordered list of blocks
type Blockchain struct {
	blocks []*Block
}

// BlockChain is a global variable that'll return the mutated Blockchain struct
var BlockChain *Blockchain

// AddBlock adds a Block to a Blockchain
func (bc *Blockchain) AddBlock(data Transaction) {
	// get previous block
	prevBlock := bc.blocks[len(bc.blocks)-1]
	// create new block
	block := CreateBlock(prevBlock, data)
	// validate integrity of blocks
	if validBlock(block, prevBlock) {
		// TODO: and if role permission okay
		bc.blocks = append(bc.blocks, block)
	}
}

func GenesisBlock() *Block {
	return CreateBlock(&Block{}, Transaction{IsGenesis: true})
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}

func validBlock(block, prevBlock *Block) bool {
	// Confirm the hashes
	if prevBlock.Hash != block.PrevHash {
		return false
	}
	// confirm the block's hash is valid
	if !block.validateHash(block.Hash) {
		return false
	}
	// Check the position to confirm its been incremented
	if prevBlock.Position+1 != block.Position {
		return false
	}
	return true
}

func validRole(userRole string) bool {
	// TODO: get user by UserID
	log.Printf("userRole: %v", userRole)
	return true
}

func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	if b.Hash != hash {
		return false
	}
	return true
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.blocks, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	// write JSON string
	io.WriteString(w, string(jbytes))
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var transaction Transaction
	if validRole(transaction.UserRole) {
		// Handle error
		if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("could not write Block: %v", err)
			w.Write([]byte("could not write block"))
			return
		}

		// Create block
		BlockChain.AddBlock(transaction)
		resp, err := json.MarshalIndent(transaction, "", " ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("could not marshal payload: %v", err)
			w.Write([]byte("could not write block"))
			return
		}

		// Response
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func newMedicalRecord(w http.ResponseWriter, r *http.Request) {
	var medicalRecord MedicalRecord
	if err := json.NewDecoder(r.Body).Decode(&medicalRecord); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not create: %v", err)
		w.Write([]byte("could not create new MedicalRecord"))
		return
	}
	// We'll create an ID, concatenating the isdb and publish date
	// This isn't an efficient way but serves for this tutorial
	h := md5.New()
	io.WriteString(h, medicalRecord.FullName+medicalRecord.CreationDate)
	medicalRecord.WalletAddress = fmt.Sprintf("%x", h.Sum(nil))

	// send back payload
	resp, err := json.MarshalIndent(medicalRecord, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not save medicalRecord data"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func main() {
	// initialize the blockchain and store in var
	BlockChain = NewBlockchain()

	// register router
	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")

	r.HandleFunc("/transaction", writeBlock).Methods("GET")
	r.HandleFunc("/transaction", writeBlock).Methods("POST")

	r.HandleFunc("/wallet", newMedicalRecord).Methods("GET")
	r.HandleFunc("/wallet", newMedicalRecord).Methods("POST")

	// TODO: new user
	// r.HandleFunc("/user", newMedicalRecord).Methods("GET")
	// r.HandleFunc("/user", newMedicalRecord).Methods("POST")

	// dump the state of the Blockchain to the console
	go func() {
		//for {
		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
		//}
	}()
	log.Println("Listening on port 3000")

	log.Fatal(http.ListenAndServe(":3000", r))
}
