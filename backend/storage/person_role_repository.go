package storage

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	supa "github.com/supabase-community/supabase-go"
)

// PersonRoleRepository provides methods to interact with the PersonRole table
type PersonRoleRepository struct {
	client *supa.Client
}

// NewPersonRoleRepository creates a new PersonRoleRepository
func NewPersonRoleRepository(client *supa.Client) *PersonRoleRepository {
	return &PersonRoleRepository{
		client: client,
	}
}

// Create adds a new person_role to the database
func (r *PersonRoleRepository) Create(ctx context.Context, personRole model.PersonRole) (*model.PersonRole, error) {
	// Ensure ID is set if it's nil
	if personRole.ID == uuid.Nil {
		personRole.ID = uuid.New()
	}

	data, count, err := r.client.From("person_role").Insert(personRole, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating person_role: %v", err)
		return nil, err
	}

	if count == 0 {
		log.Println("No rows returned after person_role insert")
		return nil, nil // Or an error indicating no result
	}

	var createdPersonRoles []model.PersonRole
	err = json.Unmarshal([]byte(data), &createdPersonRoles)
	if err != nil {
		log.Printf("Error parsing person_role data: %v", err)
		return nil, err
	}

	if len(createdPersonRoles) == 0 {
		log.Println("Parsed person_role data is empty")
		return nil, nil // Or an error
	}

	return &createdPersonRoles[0], nil
}
