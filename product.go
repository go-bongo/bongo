package bongo

import (
	// "errors"
	// "fmt"
	// "github.com/jinzhu/now"
	"labix.org/v2/mgo/bson"
	// "maxwellhealth/common"
	"maxwellhealth/common/models/types"
	"maxwellhealth/common/utilities"
	"time"
)

func NewProduct() *Product {
	prod := new(Product)

	prod.TemplateIds = &ProductTemplates{}
	prod.Images = &ProductImages{}

	prod.Info = make(map[string]interface{})
	prod.Display = make(map[string]interface{})
	prod.SelectedOptions = &ProductSelectedOptions{}
	prod.Props = &ProductProps{}
	prod.Costs = &ProductCosts{}
	prod.CalculatedValues = &ProductCalculatedValues{
		MaximumBenefit:      &types.Strategy{},
		Benefit:             &types.BenefitInfo{},
		FixedCoverageAmount: &ProductCalculatedCoverage{},
		MaximumCoverage:     &ProductCalculatedCoverage{},
		CoverageAmount:      &ProductCalculatedCoverage{},
		GuaranteedIssue:     &ProductCalculatedCoverage{},
	}
	prod.DefinedBenefit = &ProductDefinedBenefit{}
	prod.Calculation = &ProductCalculation{}
	prod.GuaranteedIssue = &ProductCalculatedCoverage{}

	return prod
}

type ProductCosts struct {

	// The total cost
	Amount float64
	// For multipliers, etc, we need the base cost to figure out the eventual amount
	Base float64

	// How often the cost is paid
	Frequency string

	// Any errors incurred while calculating the costs (deprecated?)
	Errors []string

	// The module to use to calculate the costs
	Module string

	// There can either be generic costs, or costs by ROLE (sp, ch). EE costs will always be here if they are split up, but we also have Spouse and Child costs which have their own banding
	Banding *ProductBandingConfig
	Config  *ProductCostsConfig
}

type ProductCostsConfig struct {
	SpouseRate bool `bson:"spouseRate"`
}

type ProductRegionBandConfig struct {
	Enabled           bool
	UseEmployerRegion bool
}

// This could technically just be an array of strings and the system would be ok, but we need a way to easily show all of the available options they want to band on so we can generate the cost input CSV in the portal builder. For that reason we'll store the field name + values instead of just the field name. Note that we DON'T need to do this for customBooleans because the only possible values for those are known (true/false)
type ProductSelectedOptionsBandConfig struct {
	FieldName string `bson:"fieldName"`
	Values    []string
}

type ProductBandingConfig struct {
	Age                                 bool
	InsuranceAge                        bool `bson:"insuranceAge"`
	InsuranceAgeEffectiveDate           bool `bson:"insuranceAgeEffectiveDate"`
	IndividualAge                       bool `bson:"individualAge"`
	IndividualInsuranceAge              bool `bson:"individualInsuranceAge"`
	IndividualInsuranceAgeEffectiveDate bool `bson:"individualInsuranceAgeEffectiveDate"`
	SpouseAge                           bool `bson:"spouseAge"`
	CoverageTier                        bool `bson:"coverageTier"`
	DependentCoverageTier               bool `bson:"dependentCoverageTier"`
	Gender                              bool
	Smoking                             bool
	Region                              *ProductRegionBandConfig

	// Order for these matters.
	CustomBooleans  []string                            `bson:"customBooleans"`
	SelectedOptions []*ProductSelectedOptionsBandConfig `bson:"selectedOptions"`
}

type ProductCalculation struct {
	Module string
}

type ProductTransitByMonth struct {
	Jan *ProductTransitCostPerMonth
	Feb *ProductTransitCostPerMonth
	Mar *ProductTransitCostPerMonth
	Apr *ProductTransitCostPerMonth
	May *ProductTransitCostPerMonth
	Jun *ProductTransitCostPerMonth
	Jul *ProductTransitCostPerMonth
	Aug *ProductTransitCostPerMonth
	Sep *ProductTransitCostPerMonth
	Oct *ProductTransitCostPerMonth
	Nov *ProductTransitCostPerMonth
	Dec *ProductTransitCostPerMonth
}

type ProductTransitCostPerMonth struct {
	Parking float64
	Transit float64
	Biking  float64
}

type ProductSelectedOptions struct {
	FamilyMembersToCover []bson.ObjectId
	CoverageAmount       float64
	Tax                  string
	SpouseCoverageAmount float64
	ChildCoverageAmount  float64

	CustomFields map[string]string `bson:"customFields"`

	TransitPerMonth *ProductTransitByMonth `bson:"transitCostPerMonth"`
}

type ProductDefinedBenefit struct {
	Module   string
	Strategy string
	Amount   float64
}

type ProductTemplates struct {
	EmployerSummary bson.ObjectId `bson:"employerSummary,omitempty"`
	EmployerDetails bson.ObjectId `bson:"employerDetails,omitempty"`
	EmployerOptions bson.ObjectId `bson:"employerOptions,omitempty"`
	EmployeeSummary bson.ObjectId `bson:"employeeSummary,omitempty"`
	EmployeeOptions bson.ObjectId `bson:"employeeOptions,omitempty"`
	EmployeeDetails bson.ObjectId `bson:"employeeDetails,omitempty"`
	MobileProduct   bson.ObjectId `bson:"mobileProduct,omitempty"`
	MobileWellness  bson.ObjectId `bson:"mobileWellness,omitempty"`
	MobileHelp      bson.ObjectId `bson:"mobileHelp,omitempty"`
}

type ProductCalculatedCoverage struct {
	Employee *types.Strategy
	Spouse   *types.Strategy
	Child    *types.Strategy
}

type ProductCalculatedValues struct {
	MaximumBenefit      *types.Strategy `bson:"maximumBenefit"`
	Benefit             *types.BenefitInfo
	PreTax              bool                       `bson:"preTax"`
	PostTax             bool                       `bson:"postTax"`
	FixedCoverageAmount *ProductCalculatedCoverage `bson:"fixedCoverageAmount"`
	MaximumCoverage     *ProductCalculatedCoverage `bson:"maximumCoverage"`
	CoverageAmount      *ProductCalculatedCoverage `bson:"coverageAmount"`
	GuaranteedIssue     *ProductCalculatedCoverage `bson:"guaranteedIssue"`
}

type ProductImages struct {
	Long   string
	Bundle string
	Square string
	Mobile string
}

type Product struct {
	Id bson.ObjectId `bson:"_id"`

	// References
	VendorId   bson.ObjectId `bson:"vendorId,omitempty"`
	EmployerId bson.ObjectId `bson:"employerId",omitempty`
	GroupId    bson.ObjectId `bson:"groupId",omitempty`
	FamilyId   bson.ObjectId `bson:"familyId",omitempty`

	// Basic Info
	NameLong       string `bson:"nameLong"`
	NameShort      string `bson:"nameShort"`
	Nickname       string
	PlanId         string `bson:"planId"`
	MasterPolicyId string `bson:"masterPolicyId"`
	Description    string
	InfoUrl        string            `bson:"infoUrl"`
	TemplateIds    *ProductTemplates `bson:"templateIds"`

	// Categorization
	Type     string
	SubType  string `bson:"subType"`
	Category string
	Tags     []string
	Class    string

	// Flags
	Template                bool
	Waivable                bool
	PreTax                  bool `bson:"preTax"`
	ShowPayPeriodAmount     bool `bson:"showPayPeriodAmount"`
	NotAvailableForShopping bool `bson:"notAvailableForShopping"`
	ShowFrequencyAmount     bool `bson:"showFrequencyAmount"`
	DefinedContribution     bool `bson:"definedContribution"`

	// Configuration Options
	Statuses []string
	Images   *ProductImages

	// These are just front-end configuration and won't break the system, so we can let this be an arbitrary interface map
	Info    map[string]interface{}
	Display map[string]interface{}

	// Enrollment Settings
	DateStart        time.Time `bson:"dateStart"`
	DateEnd          time.Time `bson:"dateEnd"`
	Owned            bool
	Waived           bool
	Terminated       bool
	Status           string
	SelectedOptions  *ProductSelectedOptions `bson:"selectedOptions" encrypted:"true"`
	Props            *ProductProps
	Costs            *ProductCosts            `encrypted:"true"`
	CalculatedValues *ProductCalculatedValues `bson:"calculatedValues" encrypted:"true"`
	DefinedBenefit   *ProductDefinedBenefit   `bson:"definedBenefit"`
	Calculation      *ProductCalculation
	GuaranteedIssue  *ProductCalculatedCoverage `bson:"guaranteedIssue"`
}

// Round the costs pre-save
func (p *Product) RoundCosts() {
	p.Costs.Amount = utilities.Round(p.Costs.Amount, .5, 2)
	p.CalculatedValues.MaximumBenefit.Amount = utilities.Round(p.CalculatedValues.MaximumBenefit.Amount, .5, 2)
	p.CalculatedValues.Benefit.Amount = utilities.Round(p.CalculatedValues.Benefit.Amount, .5, 2)
}

func (p *Product) BeforeSave() {
	p.RoundCosts()
}

// type ProductBandingCalculationArguments struct {
// 	Primary *User
// 	Spouse  *User

// 	// If we're doing individually banded rates we need the "individual".
// 	User     *User
// 	Family   *Family
// 	Employer *Employer
// }

// func (p *Product) GetCostBand(args *ProductBandingCalculationArguments) (*Band, error) {

// 	band := &Band{}

// 	config := p.Costs.Banding

// 	var search = bson.M{
// 		"productId": p.Id,
// 	}

// 	/////////////////
// 	/// AGE BANDING
// 	/////////////////

// 	// There can only be one "age" band
// 	var age int
// 	ageBanded := false
// 	if config.Age {
// 		ageBanded = true
// 		age = args.Primary.GetAge()
// 	} else if config.InsuranceAge {
// 		// To get insurance age we find the difference in years between the user's birthday and the same date on the product's date start year, or in other words, 1/1 on birth year to 1/1 on product start year
// 		ageBanded = true
// 		dob := now.New(args.Primary.Dob).BeginningOfYear()
// 		start := now.New(p.DateStart).BeginningOfYear()
// 		age = utilities.YearsBetween(&dob, &start)
// 	} else if config.InsuranceAgeEffectiveDate {
// 		ageBanded = true
// 		// How old will the user be on the product effective date?
// 		age = args.Primary.GetAgeOnDate(&p.DateStart)
// 	} else if config.SpouseAge {
// 		ageBanded = true
// 		age = args.Spouse.GetAge()
// 	}

// 	if ageBanded {
// 		search["age.start"] = bson.M{
// 			"$lte": age,
// 		}
// 		search["age.end"] = bson.M{
// 			"$gte": age,
// 		}
// 	}

// 	/////////////////
// 	/// REGION BANDING
// 	/////////////////
// 	if config.Region.Enabled {
// 		zip := args.Primary.Address.Zip

// 		if zip == 0 && config.Region.UseEmployerRegion {
// 			zip = args.Employer.Address.Zip
// 		}

// 		// If we still don't have a zip, error out. Costs cannot be calculated
// 		if zip == 0 {
// 			return band, errors.New("Zip is required for calculating region-banded costs. Cannot continue")
// 		}

// 		search["region.start"] = bson.M{
// 			"$lte": zip,
// 		}
// 		search["region.end"] = bson.M{
// 			"$gte": zip,
// 		}
// 	}

// 	/////////////////
// 	/// SMOKING BANDED (only for EE)
// 	/////////////////
// 	if config.Smoking {
// 		search["smoking"] = args.Primary.Smoking
// 	}

// 	/////////////////
// 	/// CUSTOM BOOLEANS
// 	/////////////////
// 	if len(config.CustomBooleans) > 0 {
// 		for index, b := range config.CustomBooleans {
// 			// What's the value?
// 			if val, ok := args.Family.CustomFields[b]; ok {
// 				search[fmt.Sprintf("customBooleans.%d", index)] = val
// 			} else {
// 				search[fmt.Sprintf("customBooleans.%d", index)] = false
// 			}
// 		}
// 	}

// 	/////////////////
// 	/// CUSTOM SELECTED OPTIONS
// 	/////////////////
// 	if len(config.SelectedOptions) > 0 {
// 		for index, opt := range config.SelectedOptions {
// 			// What's the value?
// 			if val, ok := p.SelectedOptions.CustomFields[opt.FieldName]; ok {
// 				search[fmt.Sprintf("selectedOptions.%d", index)] = val
// 			} else {
// 				search[fmt.Sprintf("selectedOptions.%d", index)] = "default"
// 			}
// 		}
// 	}

// 	// Now we finally have the search paramters. Do a search for one matching band
// 	err := common.Databases.Core.FindOne(search, band)

// 	// fmt.Println(search)
// 	return band, err
// }

// This is what's used to calculate costs, defined benefit, contributions, etc
type ProductProps struct {
	// EE/Generic Props
	BenefitFrequency              string  `bson:"benefitFrequency"`
	DeductionFrequency            string  `bson:"deductionFrequency"`
	MaximumCoverageStrategy       string  `bson:"maximumCoverageStrategy"`
	CostDivisor                   float64 `bson:"costDivisor"`
	CoverageAmountIncrement       float64 `bson:"coverageAmountIncome"`
	CostAddition                  float64 `bson:"costAddition"`
	GuaranteedIssue               float64 `bson:"guaranteedIssue"`
	MaximumCoverageAmount         float64 `bson:"maximumCoverageAmount"`
	MinimumCoverageAmount         float64 `bson:"minimumCoverageAmount"`
	RoundingNumber                int     `bson:"roundingNumber"`
	SalaryMultiplier              float64 `bson:"salaryMultiplier"`
	SalaryPercentage              float64 `bson:"salaryPercentage"`
	CoverageAmountBasedOnSalary   bool    `bson:"coverageAmountBasedOnSalary"`
	CoverageAmount                float64 `bson:"coverageAmount"`
	MaximumContribution           float64 `bson:"maximumContribution"`
	MaxBenefitStrategy            string  `bson:"maxBenefitStrategy"`
	MaxBenefitStrategyWinner      string  `bson:"maxBenefitStrategyWinner"`
	MaxBenefit                    float64 `bson:"maxBenefit"`
	EEFixedAmount                 float64 `bson:"eeFixedAmount"`
	CostStrategy                  string  `bson:"costStrategy"`
	FixedCoverageStrategyWinner   string  `bson:"fixedCoverageStrategyWinner"`
	FixedCoverageAmount           float64 `bson:"fixedCoverageAmount"`
	EOIProduct                    bool    `bson:"EOIProduct"`
	FixedCoverageStrategy         string  `bson:"fixedCoverageStrategy"`
	MaximumCoverageStrategyWinner string  `bson:"maximumCoverageStrategyWinner"`
	FixedCoverageSalaryMultiplier float64 `bson:"fixedCoverageSalaryMultiplier"`
	EeSmoking                     bool    `bson:"eeSmoking"`

	// Transit stuff
	TransitContribution bool `bson:"transitContribution"`
	BikingContribution  bool `bson:"bikingContribution"`
	ParkingContribution bool `bson:"parkingContribution"`

	// Child props
	ChildAddition                    float64 `bson:"childAddition"`
	ChildEEPercentage                float64 `bson:"childEEPercentage"`
	ChildEligible                    bool    `bson:"childEligible"`
	ChildGuaranteedIssue             float64 `bson:"childGuaranteedIssue"`
	ChildIncrement                   float64 `bson:"childIncrement"`
	ChildMinimumCoverageAmount       float64 `bson:"childMinimumCoverageAmount"`
	ChildMaximumCoverageAmount       float64 `bson:"childMaximumCoverageAmount"`
	ChildMaximumCoverageStrategy     string  `bson:"childMaximumCoverageStrategy"`
	ChildFixedCoverageStrategyWinner string  `bson:"childFixedCoverageStrategyWinner"`
	ChildFixedCoverageAmount         float64 `bson:"childFixedCoverageAmount"`
	ChildEEMultiplier                float64 `bson:"childEEMultiplier"`
	ChildRoundingNumber              int     `bson:"childRoundingNumber"`
	ChildFixedCoverageStrategy       string  `bson:"childFixedCoverageStrategy"`
	ChildFixedAmount                 float64 `bson:"childFixedAmount"`

	// Spouse props
	SpouseAddition                      float64 `bson:"spouseAddition"`
	SpouseEEPercentage                  float64 `bson:"spouseEEPercentage"`
	SpouseEligible                      bool    `bson:"spouseEligible"`
	SpouseIncrement                     float64 `bson:"spouseIncrement"`
	SpouseGuaranteedIssue               float64 `bson:"spouseGuaranteedIssue"`
	SpouseMinimumCoverageAmount         float64 `bson:"spouseMinimumCoverageAmount"`
	SpouseMaximumCoverageAmount         float64 `bson:"spouseMaximumCoverageAmount"`
	SpouseMaximumCoverageStrategy       string  `bson:"spouseMaximumCoverageStrategy"`
	SpouseFixedCoverageStrategyWinner   string  `bson:"spouseFixedCoverageStrategyWinner"`
	SpouseFixedCoverageAmount           float64 `bson:"spouseFixedCoverageAmount"`
	SpouseEEMultiplier                  float64 `bson:"spouseEEMultiplier"`
	SpouseRoundingNumber                int     `bson:"spouseRoundingNumber"`
	SpouseFixedCoverageStrategy         string  `bson:"spouseFixedCoverageStrategy"`
	SpouseMaximumCoverageStrategyWinner string  `bson:"spouseMaximumCoverageStrategyWinner"`
	SpouseFixedAmount                   float64 `bson:"spouseFixedAmount"`
}

// public $showPayPeriodAmount = true;
// public $showFrequencyAmount = true;
// public $templateVars;
// public $statuses = [
//     'open',
//     'toBeReviewed',
//     'selected',
//     'confirmed',
//     'processing',
//     'processing_edi',
//     'closed'
// ];
// public $images = [
//     'long' => '',
//     'bundle' => '',
//     'square' => '',
//     'mobile' => ''
// ];
// public $info = [];
// public $display = [];
// public $props = [];
// public $advancedProps = [];
// public $optionsDialog = [];
// public $payrollModule = 'BasePayrollModule';
// public $costs = [
//     'module' => 'BaseCostModule'
// ];
// public $price = [];
// public $definedBenefit = [];
// public $contributionStatement = false;
// public $showContribution = true;
// public $beneficiariesPrompt = false;
// public $definedContribution = false;
// public $edi = false;
// protected $dynamicAttributes = [ 'vendor' ];
// public $ediType = false;
