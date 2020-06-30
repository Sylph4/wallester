package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/vcraescu/go-paginator"
	"github.com/vcraescu/go-paginator/adapter"
	"github.com/vcraescu/go-paginator/view"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var db *sql.DB
var tpl *template.Template

//Email validation regexp
var re = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "admin"
	dbname   = "postgres"
)

type Customer struct {
	ID        int
	FirstName string
	LastName  string
	BirthDate string
	Gender    string
	Email     string
	Address   string
}

type SearchData struct {
	Customers       []Customer
	Pages           []int
	CurrentPage     int
	SearchParameter string
}

func init() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Сonnected to database.")

	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func main() {

	http.HandleFunc("/", index)
	http.HandleFunc("/customers", showCustomers)
	http.HandleFunc("/editcustomer", editCustomer)
	http.HandleFunc("/search", searchCustomer)
	http.HandleFunc("/editcustomerprocess", editCustomerProcess)
	http.HandleFunc("/createcustomer", createCustomerForm)
	http.HandleFunc("/createcustomerprocess", createCustomerProcess)
	http.ListenAndServe(":8080", nil)

}

func searchCustomer(w http.ResponseWriter, r *http.Request) {

	page, err := strconv.Atoi(r.FormValue("page"))
	searchString := r.FormValue("parameter")

	if err != nil || searchString == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	words := strings.Fields(searchString)

	rows, err := db.Query("SELECT * FROM Customers WHERE first_name=$1 OR last_name=$2", words[0], words[1])
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	customers := make([]Customer, 0)

	for rows.Next() {
		customer := Customer{}
		err := rows.Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.BirthDate, &customer.Gender, &customer.Email, &customer.Address)
		customer.BirthDate = customer.BirthDate[0:10]
		if err != nil {
			panic(err)
		}
		customers = append(customers, customer)
	}
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	p := paginator.New(adapter.NewSliceAdapter(customers), 10)

	p.SetPage(page)

	view := view.New(&p)

	if view.Current()*10 > len(customers) {
		customers = customers[view.Prev()*10 : view.Current()*10-(view.Current()*10-len(customers))]
	} else {
		customers = customers[view.Prev()*10 : view.Current()*10]
	}

	for {
		if customers[len(customers)-1].ID == 0 {
			customers = customers[:len(customers)-1]
			continue
		}

		if customers[len(customers)-1].ID != 0 {
			break
		}

	}

	data := &SearchData{
		Customers:       customers,
		Pages:           view.Pages(),
		CurrentPage:     view.Current(),
		SearchParameter: searchString,
	}

	tpl.ExecuteTemplate(w, "search.gohtml", data)
}

func editCustomerProcess(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	customer := Customer{}
	id := r.FormValue("ID")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	customer.ID, _ = strconv.Atoi(id)
	customer.FirstName = r.FormValue("firstName")
	customer.LastName = r.FormValue("lastName")
	customer.BirthDate = r.FormValue("birthDate")
	customer.Gender = r.FormValue("gender")
	customer.Email = r.FormValue("email")
	customer.Address = r.FormValue("address")

	// validate form values
	if isFormError(&customer.FirstName, &customer.LastName, &customer.Gender, &customer.Address, &customer.Email) {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
	}

	txn, err1 := db.Begin()
	if err1 != nil {
		return
	}

	defer func() {
		// Rollback in case of Failure
		if r := recover(); r != nil {
			txn.Rollback()
		}
	}()

	// insert values
	_, err := txn.Exec("UPDATE Customers SET first_name=$1, last_name=$2, birth_date=$3, gender=$4, email=$5, address=$6 WHERE ID=$7",
		customer.FirstName, customer.LastName, customer.BirthDate, customer.Gender, customer.Email, customer.Address, customer.ID)
	if err != nil {
		txn.Rollback()
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// Commit
	err1 = txn.Commit()
	if err1 != nil {
		log.Fatal(err1)
	}

	// confirm insertion
	http.Redirect(w, r, "customers", http.StatusSeeOther)
}

func createCustomerProcess(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	customer := Customer{}
	customer.FirstName = r.FormValue("firstName")
	customer.LastName = r.FormValue("lastName")
	customer.BirthDate = r.FormValue("birthDate")
	customer.Gender = r.FormValue("gender")
	customer.Email = r.FormValue("email")
	customer.Address = r.FormValue("address")

	// validate form values
	if isFormError(&customer.FirstName, &customer.LastName, &customer.Gender, &customer.Address, &customer.Email) {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
	}

	// insert values
	_, err := db.Exec("INSERT INTO Customers(first_name, last_name, birth_date, gender, email, address)  VALUES ($1, $2, $3, $4, $5, $6)",
		customer.FirstName, customer.LastName, customer.BirthDate, customer.Gender, customer.Email, customer.Address)

	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// confirm insertion
	http.Redirect(w, r, "customers", http.StatusSeeOther)
}

func showCustomers(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM Customers")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	customers := make([]Customer, 0)

	for rows.Next() {
		customer := Customer{}
		err := rows.Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.BirthDate, &customer.Gender, &customer.Email, &customer.Address)
		customer.BirthDate = customer.BirthDate[0:10]
		if err != nil {
			panic(err)
		}
		customers = append(customers, customer)
	}
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "all.gohtml", customers)

}

func editCustomer(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM Customers WHERE id = $1", id)

	customer := Customer{}
	err := row.Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.BirthDate, &customer.Gender, &customer.Email, &customer.Address)
	customer.BirthDate = customer.BirthDate[0:10]
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "update.gohtml", customer)

}

func createCustomerForm(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "create.gohtml", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "customers", http.StatusSeeOther)
}

func isFormError(firstName, lastName, gender, address, email *string) bool {
	if (*firstName == "" || len(*firstName) > 100) ||
		(*lastName == "" || len(*lastName) > 100) ||
		(*gender != "Male" && *gender != "Female") ||
		(*address == "" || len(*address) > 200) ||
		(!re.MatchString(*email)) {
		return true
	} else {
		return false
	}
}
