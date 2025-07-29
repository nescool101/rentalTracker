package storage

import (
	supa "github.com/supabase-community/supabase-go"
)

// RepositoryFactory creates and manages repository instances
type RepositoryFactory struct {
	client                       *supa.Client
	personRepository             *PersonRepository
	propertyRepository           *PropertyRepository
	rentalRepository             *RentalRepository
	userRepository               *UserRepository
	rentPaymentRepository        *RentPaymentRepository
	rentalHistoryRepository      *RentalHistoryRepository
	maintenanceRequestRepository *MaintenanceRequestRepository
	pricingRepository            *PricingRepository
	contractSigningRepository    *ContractSigningRepository
	bankAccountRepository        *BankAccountRepository
	personRoleRepository         *PersonRoleRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(client *supa.Client) *RepositoryFactory {
	return &RepositoryFactory{
		client: client,
	}
}

// GetPersonRepository returns a person repository instance
func (f *RepositoryFactory) GetPersonRepository() *PersonRepository {
	if f.personRepository == nil {
		f.personRepository = NewPersonRepository(f.client)
	}
	return f.personRepository
}

// GetPropertyRepository returns a property repository instance
func (f *RepositoryFactory) GetPropertyRepository() *PropertyRepository {
	if f.propertyRepository == nil {
		f.propertyRepository = NewPropertyRepository(f.client)
	}
	return f.propertyRepository
}

// GetRentalRepository returns a rental repository instance
func (f *RepositoryFactory) GetRentalRepository() *RentalRepository {
	if f.rentalRepository == nil {
		f.rentalRepository = NewRentalRepository(f.client)
	}
	return f.rentalRepository
}

// GetUserRepository returns a user repository instance
func (f *RepositoryFactory) GetUserRepository() *UserRepository {
	if f.userRepository == nil {
		f.userRepository = NewUserRepository(f.client)
	}
	return f.userRepository
}

// GetRentPaymentRepository returns a rent payment repository instance
func (f *RepositoryFactory) GetRentPaymentRepository() *RentPaymentRepository {
	if f.rentPaymentRepository == nil {
		f.rentPaymentRepository = NewRentPaymentRepository(f.client)
	}
	return f.rentPaymentRepository
}

// GetRentalHistoryRepository returns a rental history repository instance
func (f *RepositoryFactory) GetRentalHistoryRepository() *RentalHistoryRepository {
	if f.rentalHistoryRepository == nil {
		f.rentalHistoryRepository = NewRentalHistoryRepository(f.client)
	}
	return f.rentalHistoryRepository
}

// GetMaintenanceRequestRepository returns a maintenance request repository instance
func (f *RepositoryFactory) GetMaintenanceRequestRepository() *MaintenanceRequestRepository {
	if f.maintenanceRequestRepository == nil {
		f.maintenanceRequestRepository = NewMaintenanceRequestRepository(f.client)
	}
	return f.maintenanceRequestRepository
}

// GetPricingRepository returns a pricing repository instance
func (f *RepositoryFactory) GetPricingRepository() *PricingRepository {
	if f.pricingRepository == nil {
		f.pricingRepository = NewPricingRepository(f.client)
	}
	return f.pricingRepository
}

// GetBankAccountRepository returns a bank account repository instance
func (f *RepositoryFactory) GetBankAccountRepository() *BankAccountRepository {
	if f.bankAccountRepository == nil {
		f.bankAccountRepository = NewBankAccountRepository(f.client)
	}
	return f.bankAccountRepository
}

// GetContractSigningRepository returns a contract signing repository instance
func (f *RepositoryFactory) GetContractSigningRepository() *ContractSigningRepository {
	if f.contractSigningRepository == nil {
		f.contractSigningRepository = NewContractSigningRepository(f.client)
	}
	return f.contractSigningRepository
}

// GetPersonRoleRepository returns a person role repository instance
func (f *RepositoryFactory) GetPersonRoleRepository() *PersonRoleRepository {
	if f.personRoleRepository == nil {
		f.personRoleRepository = NewPersonRoleRepository(f.client)
	}
	return f.personRoleRepository
}

// GetClient returns the underlying Supabase client
func (f *RepositoryFactory) GetClient() *supa.Client {
	return f.client
}
