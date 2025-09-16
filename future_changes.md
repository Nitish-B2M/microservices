### make change in product models, add these below field according to requirements
```type Product struct {
    ImageURL      string     `json:"image_url"`      // Default ""
    SKU            string     `json:"sku"`            // Default ""
    Manufacturer  string     `json:"manufacturer"`  // Default ""
    Brand          string     `json:"brand"`          // Default ""
    ReviewCount   int         `json:"review_count"`   // Default 0
    Weight        float64    `json:"weight"`         // Default 0.0
    Dimensions    string     `json:"dimensions"`     // Default ""
} 
```

## Add more apis to product 
1. filter out category
2. add new category(admin only)
3. retrieve category 
4. Media upload

## interesting topic
- if we want to automatically insert into 3rd table entries of table1 and table2 data then use 'gorm:"many2many:product_tags;"' into struct
- more details step
	+ type Product struct {
	+	Tags      []Tag     `json:"tags" gorm:"many2many:product_tags;"`
	+ }
	+ here we have 3 table, table1 = product, table2 = tags, and table3 = product_tags
	+ so when we have table1 product_id=123 and table2 tag_id=23 then it automatically add it into table3 product_tags like (product_id, tag_id)(123, 23)

- implement role based authorization or permission based authorization
  + user?.role?.active_role?.permissions?.includes('manage_products') future implementation