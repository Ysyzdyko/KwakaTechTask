package parser

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"menu-parser/internal/domain/entity"
	"menu-parser/internal/domain/service"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type sheetsParser struct {
	service *sheets.Service
}

func NewSheetsParser(credentialsPath string) (service.SheetsParser, error) {
	ctx := context.Background()

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets service: %w", err)
	}

	return &sheetsParser{
		service: srv,
	}, nil
}

func (p *sheetsParser) ParseMenu(ctx context.Context, spreadsheetID, restaurantName string) (*entity.Menu, error) {
	// Get spreadsheet metadata to find the first sheet name
	spreadsheet, err := p.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get spreadsheet metadata: %w", err)
	}

	if len(spreadsheet.Sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in spreadsheet")
	}

	// Use the first sheet
	sheetName := spreadsheet.Sheets[0].Properties.Title
	
	// Try different range formats
	var resp *sheets.ValueRange
	var readRange string
	var lastErr error
	
	// Try format: "SheetName!A:Z" (if sheet name has spaces, use single quotes)
	if strings.Contains(sheetName, " ") || strings.Contains(sheetName, "'") {
		readRange = fmt.Sprintf("'%s'!A:Z", strings.ReplaceAll(sheetName, "'", "''"))
	} else {
		readRange = fmt.Sprintf("%s!A:Z", sheetName)
	}
	resp, err = p.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		lastErr = err
		// Try format without sheet name (uses first sheet by default)
		readRange = "A:Z"
		resp, err = p.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
		if err != nil {
			// Try with just the sheet name (gets all data)
			readRange = sheetName
			resp, err = p.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve data from sheet (tried ranges: %s, A:Z, %s): %w (last error: %v)", 
					fmt.Sprintf("%s!A:Z", sheetName), sheetName, err, lastErr)
			}
		}
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("no data found in spreadsheet")
	}

	menu := &entity.Menu{
		Name:             restaurantName,
		RestaurantID:     restaurantName,
		Products:         []entity.Product{},
		AttributesGroups: []entity.AttributesGroup{},
		Attributes:       []entity.Attribute{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	products, attributesGroups, attributes := p.parseSheetData(resp.Values)

	menu.Products = products
	menu.AttributesGroups = attributesGroups
	menu.Attributes = attributes

	return menu, nil
}

func (p *sheetsParser) parseSheetData(rows [][]interface{}) ([]entity.Product, []entity.AttributesGroup, []entity.Attribute) {
	var products []entity.Product
	var attributesGroups []entity.AttributesGroup
	var attributes []entity.Attribute

	currentProduct := ""
	currentPrice := 0.0
	currentPriceOld := 0.0
	currentAttributes := []string{}
	productExtID := 1001000

	attributeMap := make(map[string]bool)
	attributesGroupMap := make(map[string]*entity.AttributesGroup)

	for i, row := range rows {
		if len(row) < 2 {
			continue
		}

		if len(row) > 1 && row[1] != nil && row[1] != "" {
			productName := strings.TrimSpace(fmt.Sprintf("%v", row[1]))

			if productName == "Glovo" || productName == "" {
				continue
			}

			if currentProduct != "" {
				product := entity.Product{
					ExtID:    strconv.Itoa(productExtID),
					Name:     currentProduct,
					Price:    currentPrice,
					PriceOld: currentPriceOld,
					Status:   string(entity.ProductStatusAvailable),
				}
				if len(currentAttributes) > 0 {
					product.Attributes = map[string]interface{}{
						"options": currentAttributes,
					}
				}
				products = append(products, product)
				productExtID++
			}

			currentProduct = productName
			currentAttributes = []string{}

			if len(row) > 3 && row[3] != nil {
				if price, err := parsePrice(fmt.Sprintf("%v", row[3])); err == nil {
					currentPrice = price
				}
			}
			if len(row) > 4 && row[4] != nil {
				if price, err := parsePrice(fmt.Sprintf("%v", row[4])); err == nil {
					currentPriceOld = price
				}
			}
		}

		if len(row) > 7 && row[7] != nil {
			attrValue := strings.TrimSpace(fmt.Sprintf("%v", row[7]))
			if attrValue != "" && attrValue != currentProduct {
				currentAttributes = append(currentAttributes, attrValue)

				if !attributeMap[attrValue] {
					attributeMap[attrValue] = true
					attributes = append(attributes, entity.Attribute{
						ID:   fmt.Sprintf("attr_%d", len(attributes)),
						Name: attrValue,
					})
				}
			}
		}

		if len(row) > 1 && row[1] != nil {
			categoryName := strings.TrimSpace(fmt.Sprintf("%v", row[1]))
			if categoryName != "" && categoryName != currentProduct &&
				(strings.Contains(categoryName, "предложения") ||
					strings.Contains(categoryName, "позиции") ||
					strings.Contains(categoryName, "Glovo")) {
				if _, exists := attributesGroupMap[categoryName]; !exists {
					group := &entity.AttributesGroup{
						ID:         fmt.Sprintf("group_%d", len(attributesGroups)),
						Name:       categoryName,
						Attributes: []entity.Attribute{},
						IsRequired: false,
					}
					attributesGroups = append(attributesGroups, *group)
					attributesGroupMap[categoryName] = group
				}
			}
		}

		if i > 0 && len(row) > 1 && row[1] == nil && currentProduct != "" {
			if i+1 < len(rows) && len(rows[i+1]) > 1 && rows[i+1][1] == nil {
				if currentProduct != "" {
					product := entity.Product{
						ExtID:    strconv.Itoa(productExtID),
						Name:     currentProduct,
						Price:    currentPrice,
						PriceOld: currentPriceOld,
						Status:   string(entity.ProductStatusAvailable),
					}
					if len(currentAttributes) > 0 {
						product.Attributes = map[string]interface{}{
							"options": currentAttributes,
						}
					}
					products = append(products, product)
					productExtID++
					currentProduct = ""
					currentAttributes = []string{}
				}
			}
		}
	}

	if currentProduct != "" {
		product := entity.Product{
			ExtID:    strconv.Itoa(productExtID),
			Name:     currentProduct,
			Price:    currentPrice,
			PriceOld: currentPriceOld,
			Status:   string(entity.ProductStatusAvailable),
		}
		if len(currentAttributes) > 0 {
			product.Attributes = map[string]interface{}{
				"options": currentAttributes,
			}
		}
		products = append(products, product)
	}

	return products, attributesGroups, attributes
}

func parsePrice(priceStr string) (float64, error) {
	priceStr = strings.ReplaceAll(priceStr, " ", "")
	priceStr = strings.ReplaceAll(priceStr, ",", ".")
	return strconv.ParseFloat(priceStr, 64)
}

