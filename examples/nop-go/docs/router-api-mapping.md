根据前一个会话中所有已读取和分析的控制器文件，以下是按微服务模块组织的完整路由提取。前端控制器已完全分析（26个文件），并且已完全读取10个管理控制器，同时通过代码模式分析推断出其余管理控制器的路由。

---

# nopCommerce 4.90.4 Controller Route API 完整提取 (按微服务模块分组)

---

## 1. 商品服务 (catalog)

### 前端 - ProductController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /product/{productId} | ProductDetails | int productId | View |
| POST | /product/estimateshipping | EstimateShipping | EstimateShippingModel model | Json |
| GET | /product/getproductcombinations | GetProductCombinations | int productId | Json |
| GET | /recentlyviewedproducts | RecentlyViewedProducts | - | View |
| POST | /product/productreviewsadd | ProductReviewsAdd | ProductReviewsModel model | View/Redirect |
| POST | /product/setproductreviewhelpfulness | SetProductReviewHelpfulness | int productReviewId, bool washelpful | Json |
| GET | /customer/productreviews | CustomerProductReviews | int? pageNumber | View |
| GET | /product/emailafriend/{productId} | ProductEmailAFriend | int productId | View |
| POST | /product/emailafriend | ProductEmailAFriendSend | ProductEmailAFriendModel model | View |
| POST | /product/addtocomparelist | AddProductToCompareList | int productId | Json |
| GET | /compareproducts/removeproductfromcomparelist | RemoveProductFromCompareList | int productId | Redirect |
| GET | /compareproducts | CompareProducts | - | View |
| GET | /compareproducts/clearcomparelist | ClearCompareList | - | Redirect |

### 前端 - CatalogController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /category/{categoryId} | Category | int categoryId, CatalogPagingFilteringModel command | View |
| POST | /category/getcategoryproducts | GetCategoryProducts | int categoryId, CatalogPagingFilteringModel command | PartialView |
| GET | /manufacturer/{manufacturerId} | Manufacturer | int manufacturerId, CatalogPagingFilteringModel command | View |
| POST | /manufacturer/getmanufacturerproducts | GetManufacturerProducts | int manufacturerId, CatalogPagingFilteringModel command | PartialView |
| GET | /manufacturer/all | ManufacturerAll | - | View |
| GET | /vendor/{vendorId} | Vendor | int vendorId, CatalogPagingFilteringModel command | View |
| POST | /vendor/getvendorproducts | GetVendorProducts | int vendorId, CatalogPagingFilteringModel command | PartialView |
| GET | /vendorreviews/{vendorId} | VendorReviews | int vendorId | View |
| GET | /vendor/all | VendorAll | - | View |
| GET | /producttag/{productTagId} | ProductsByTag | int productTagId, CatalogPagingFilteringModel command | View |
| POST | /producttag/gettagproducts | GetTagProducts | int productTagId, CatalogPagingFilteringModel command | PartialView |
| GET | /producttag/all | ProductTagsAll | - | View |
| GET | /newproducts | NewProducts | CatalogPagingFilteringModel command | View |
| POST | /newproducts/getnewproducts | GetNewProducts | CatalogPagingFilteringModel command | PartialView |
| GET | /newproducts/rss | NewProductsRss | - | RssActionResult |
| GET | /search | Search | SearchModel model | View |
| GET | /search/searchtermautocomplete | SearchTermAutoComplete | string term | Json |
| POST | /search/searchproducts | SearchProducts | SearchModel model | PartialView |
| GET | /search/getfilterlevelvalues | GetFilterLevelValues | int filterLevelId | Json |
| GET | /search/searchbyfilterlevelvalues | SearchByFilterLevelValues | int filterLevelId, int filterLevelValueId | View |
| POST | /search/searchproductsbyfilterlevelvalues | SearchProductsByFilterLevelValues | int filterLevelId, int filterLevelValueId | PartialView |

### 前端 - BackInStockSubscriptionController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /backinstocksubscription/subscribepopup | SubscribePopup | int productId | PartialView |
| POST | /backinstocksubscription/subscribepopup | SubscribePopupPOST | int productId | Json/OkResult |
| GET | /backinstocksubscription/customersubscriptions | CustomerSubscriptions | int? pageNumber | View |
| POST | /backinstocksubscription/customersubscriptions | CustomerSubscriptionsPOST | - | Redirect |

### Admin - ProductController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Product | Index | - | Redirect to List |
| GET | /Admin/Product/List | List | - | View |
| GET | /Admin/Product/BulkEdit | BulkEdit | - | View |
| POST | /Admin/Product/BulkEditSave | BulkEditSave | - | Redirect |
| POST | /Admin/Product/ProductList | ProductList | ProductSearchModel searchModel | Json |
| POST | /Admin/Product/BulkEditProducts | BulkEditProducts | BulkEditSearchModel searchModel | Json |
| POST | /Admin/Product/BulkEditNewProduct | BulkEditNewProduct | - | Json |
| POST | /Admin/Product/GoToSku | GoToSku | string sku | Redirect |
| GET | /Admin/Product/Create | Create | - | View |
| POST | /Admin/Product/Create | Create | ProductModel model, bool continueEditing | Redirect |
| POST | /Admin/Product/PreTranslate | PreTranslate | int id | Json |
| GET | /Admin/Product/Edit/{id} | Edit | int id | View |
| POST | /Admin/Product/Edit | Edit | ProductModel model, bool continueEditing | Redirect |
| POST | /Admin/Product/Delete | Delete | int id | Redirect |
| POST | /Admin/Product/DeleteSelected | DeleteSelected | ICollection<int> selectedIds | Json |
| POST | /Admin/Product/CopyProduct | CopyProduct | ProductModel model | Redirect |
| GET | /Admin/Product/SkuReservedWarning | SkuReservedWarning | string sku, int id | Json |
| GET | /Admin/Product/CustomersDateOfBirthDisabledWarning | CustomersDateOfBirthDisabledWarning | - | Json |
| GET | /Admin/Product/FullDescriptionGeneratorPopup | FullDescriptionGeneratorPopup | int id | View |
| POST | /Admin/Product/FullDescriptionGeneratorPopup | FullDescriptionGeneratorPopup | int id, string text | Json |
| POST | /Admin/Product/LoadProductFriendlyNames | LoadProductFriendlyNames | string productIds | Json |
| GET | /Admin/Product/RequiredProductAddPopup | RequiredProductAddPopup | - | View |
| POST | /Admin/Product/RequiredProductAddPopup | RequiredProductAddPopup | - | Json(list) |
| POST | /Admin/Product/RelatedProductList | RelatedProductList | RelatedProductSearchModel searchModel | Json |
| POST | /Admin/Product/RelatedProductUpdate | RelatedProductUpdate | RelatedProductModel model | NullJson |
| POST | /Admin/Product/RelatedProductDelete | RelatedProductDelete | int id | NullJson |
| GET | /Admin/Product/RelatedProductAddPopup | RelatedProductAddPopup | - | View |
| POST | /Admin/Product/RelatedProductAddPopup | RelatedProductAddPopup | - | Json(list) |
| POST | /Admin/Product/CrossSellProductList | CrossSellProductList | CrossSellProductSearchModel searchModel | Json |
| POST | /Admin/Product/CrossSellProductDelete | CrossSellProductDelete | int id | NullJson |
| GET | /Admin/Product/CrossSellProductAddPopup | CrossSellProductAddPopup | - | View |
| POST | /Admin/Product/CrossSellProductAddPopup | CrossSellProductAddPopup | - | Json(list) |
| POST | /Admin/Product/FilterLevelValueList | FilterLevelValueList | FilterLevelValueSearchModel searchModel | Json |
| POST | /Admin/Product/FilterLevelValueDelete | FilterLevelValueDelete | int id | NullJson |
| GET | /Admin/Product/FilterLevelValuesAddPopup | FilterLevelValuesAddPopup | - | View |
| POST | /Admin/Product/FilterLevelValuesAddPopup | FilterLevelValuesAddPopup | - | Json(list) |
| POST | /Admin/Product/AssociatedProductList | AssociatedProductList | AssociatedProductSearchModel searchModel | Json |
| POST | /Admin/Product/AssociatedProductUpdate | AssociatedProductUpdate | AssociatedProductModel model | NullJson |
| POST | /Admin/Product/AssociatedProductDelete | AssociatedProductDelete | int id | NullJson |
| GET | /Admin/Product/AssociatedProductAddPopup | AssociatedProductAddPopup | - | View |
| POST | /Admin/Product/AssociatedProductAddPopup | AssociatedProductAddPopup | - | Json(list) |
| POST | /Admin/Product/ProductPictureList | ProductPictureList | ProductPictureSearchModel searchModel | Json |
| POST | /Admin/Product/ProductPictureAdd | ProductPictureAdd | int productId, int pictureId, int displayOrder | NullJson |
| POST | /Admin/Product/ProductPictureUpdate | ProductPictureUpdate | ProductPictureModel model | NullJson |
| POST | /Admin/Product/ProductPictureDelete | ProductPictureDelete | int id | NullJson |
| GET | /Admin/Product/ProductPictureAddPopup | ProductPictureAddPopup | - | View |
| POST | /Admin/Product/ProductPictureAddPopup | ProductPictureAddPopup | - | Json(list) |
| POST | /Admin/Product/ProductAttributeMappingList | ProductAttributeMappingList | ProductAttributeMappingSearchModel searchModel | Json |
| POST | /Admin/Product/ProductAttributeMappingUpdate | ProductAttributeMappingUpdate | ProductAttributeMappingModel model | NullJson |
| POST | /Admin/Product/ProductAttributeMappingDelete | ProductAttributeMappingDelete | int id | NullJson |
| GET | /Admin/Product/ProductAttributeMappingCreate | ProductAttributeMappingCreate | int productId | View |
| POST | /Admin/Product/ProductAttributeMappingCreate | ProductAttributeMappingCreate | ProductAttributeMappingModel model | Redirect |
| GET | /Admin/Product/ProductAttributeMappingEdit | ProductAttributeMappingEdit | int id | View |
| POST | /Admin/Product/ProductAttributeMappingEdit | ProductAttributeMappingEdit | ProductAttributeMappingModel model | Redirect |
| POST | /Admin/Product/ProductAttributeValueList | ProductAttributeValueList | ProductAttributeValueSearchModel searchModel | Json |
| POST | /Admin/Product/ProductAttributeValueUpdate | ProductAttributeValueUpdate | ProductAttributeValueModel model | NullJson |
| POST | /Admin/Product/ProductAttributeValueDelete | ProductAttributeValueDelete | int id | NullJson |
| GET | /Admin/Product/ProductAttributeValueCreatePopup | ProductAttributeValueCreatePopup | int productAttributeMappingId | View |
| POST | /Admin/Product/ProductAttributeValueCreatePopup | ProductAttributeValueCreatePopup | ProductAttributeValueModel model | Json |
| GET | /Admin/Product/ProductAttributeValueEditPopup | ProductAttributeValueEditPopup | int id | View |
| POST | /Admin/Product/ProductAttributeValueEditPopup | ProductAttributeValueEditPopup | ProductAttributeValueModel model | Json |
| POST | /Admin/Product/ProductAttributeCombinationList | ProductAttributeCombinationList | ProductAttributeCombinationSearchModel searchModel | Json |
| POST | /Admin/Product/ProductAttributeCombinationUpdate | ProductAttributeCombinationUpdate | ProductAttributeCombinationModel model | NullJson |
| POST | /Admin/Product/ProductAttributeCombinationDelete | ProductAttributeCombinationDelete | int id | NullJson |
| GET | /Admin/Product/ProductAttributeCombinationCreatePopup | ProductAttributeCombinationCreatePopup | int productId | View |
| POST | /Admin/Product/ProductAttributeCombinationCreatePopup | ProductAttributeCombinationCreatePopup | ProductAttributeCombinationModel model | Json |
| GET | /Admin/Product/ProductAttributeCombinationEditPopup | ProductAttributeCombinationEditPopup | int id | View |
| POST | /Admin/Product/ProductAttributeCombinationEditPopup | ProductAttributeCombinationEditPopup | ProductAttributeCombinationModel model | Json |
| POST | /Admin/Product/GenerateAllAttributeCombinations | GenerateAllAttributeCombinations | int productId | Json |
| GET | /Admin/Product/Inventory | Inventory | int id | View |
| POST | /Admin/Product/Inventory | Inventory | ProductInventoryModel model | Redirect |
| POST | /Admin/Product/Download | Download | int id | Redirect |

### Admin - CategoryController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Category | Index | - | Redirect to List |
| GET | /Admin/Category/List | List | - | View |
| POST | /Admin/Category/List | CategoryList | CategorySearchModel searchModel | Json |
| GET | /Admin/Category/Create | Create | - | View |
| POST | /Admin/Category/Create | Create | CategoryModel model, bool continueEditing | Redirect |
| GET | /Admin/Category/Edit/{id} | Edit | int id | View |
| POST | /Admin/Category/Edit | Edit | CategoryModel model, bool continueEditing | Redirect |
| POST | /Admin/Category/Delete | Delete | int id | Redirect |
| GET | /Admin/Category/Tree | Tree | - | View |
| POST | /Admin/Category/Tree | Tree | CategoryTreeModel model | View |

### Admin - ManufacturerController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Manufacturer | Index | - | Redirect to List |
| GET | /Admin/Manufacturer/List | List | - | View |
| POST | /Admin/Manufacturer/List | ManufacturerList | ManufacturerSearchModel searchModel | Json |
| GET | /Admin/Manufacturer/Create | Create | - | View |
| POST | /Admin/Manufacturer/Create | Create | ManufacturerModel model, bool continueEditing | Redirect |
| GET | /Admin/Manufacturer/Edit/{id} | Edit | int id | View |
| POST | /Admin/Manufacturer/Edit | Edit | ManufacturerModel model, bool continueEditing | Redirect |
| POST | /Admin/Manufacturer/Delete | Delete | int id | Redirect |

### Admin - ProductReviewController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Review | Index | - | Redirect to List |
| GET | /Admin/Review/List | List | - | View |
| POST | /Admin/Review/List | ReviewList | ProductReviewSearchModel searchModel | Json |
| POST | /Admin/Review/Update | ReviewUpdate | ProductReviewModel model | NullJson |
| POST | /Admin/Review/Delete | Delete | int id | NullJson |
| POST | /Admin/Review/DeleteSelected | DeleteSelected | ICollection<int> selectedIds | Json |
| POST | /Admin/Review/ApproveSelected | ApproveSelected | ICollection<int> selectedIds | Json |
| POST | /Admin/Review/DisapproveSelected | DisapproveSelected | ICollection<int> selectedIds | Json |

### Admin - CheckoutAttributeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/CheckoutAttribute | Index | - | Redirect to List |
| GET | /Admin/CheckoutAttribute/List | List | - | View |
| POST | /Admin/CheckoutAttribute/List | CheckoutAttributeList | CheckoutAttributeSearchModel searchModel | Json |
| GET | /Admin/CheckoutAttribute/Create | Create | - | View |
| POST | /Admin/CheckoutAttribute/Create | Create | CheckoutAttributeModel model, bool continueEditing | Redirect |
| GET | /Admin/CheckoutAttribute/Edit/{id} | Edit | int id | View |
| POST | /Admin/CheckoutAttribute/Edit | Edit | CheckoutAttributeModel model, bool continueEditing | Redirect |
| POST | /Admin/CheckoutAttribute/Delete | Delete | int id | Redirect |
| POST | /Admin/CheckoutAttribute/ValueList | ValueList | CheckoutAttributeValueSearchModel searchModel | Json |
| POST | /Admin/CheckoutAttributeValue/ValueUpdate | ValueUpdate | CheckoutAttributeValueModel model | NullJson |
| POST | /Admin/CheckoutAttributeValue/ValueDelete | ValueDelete | int id | NullJson |
| GET | /Admin/CheckoutAttribute/ValueCreatePopup | ValueCreatePopup | int checkoutAttributeId | View |
| POST | /Admin/CheckoutAttribute/ValueCreatePopup | ValueCreatePopup | CheckoutAttributeValueModel model | Json |
| GET | /Admin/CheckoutAttribute/ValueEditPopup | ValueEditPopup | int id | View |
| POST | /Admin/CheckoutAttribute/ValueEditPopup | ValueEditPopup | CheckoutAttributeValueModel model | Json |

### Admin - ProductAttributeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/ProductAttribute | Index | - | Redirect to List |
| GET | /Admin/ProductAttribute/List | List | - | View |
| POST | /Admin/ProductAttribute/List | ProductAttributeList | ProductAttributeSearchModel searchModel | Json |
| GET | /Admin/ProductAttribute/Create | Create | - | View |
| POST | /Admin/ProductAttribute/Create | Create | ProductAttributeModel model, bool continueEditing | Redirect |
| GET | /Admin/ProductAttribute/Edit/{id} | Edit | int id | View |
| POST | /Admin/ProductAttribute/Edit | Edit | ProductAttributeModel model, bool continueEditing | Redirect |
| POST | /Admin/ProductAttribute/Delete | Delete | int id | Redirect |
| POST | /Admin/ProductAttribute/ValueList | ValueList | ProductAttributeValueSearchModel searchModel | Json |
| POST | /Admin/ProductAttribute/ValueUpdate | ValueUpdate | ProductAttributeValueModel model | NullJson |
| POST | /Admin/ProductAttribute/ValueDelete | ValueDelete | int id | NullJson |
| GET | /Admin/ProductAttribute/ValueCreatePopup | ValueCreatePopup | int productAttributeId | View |
| POST | /Admin/ProductAttribute/ValueCreatePopup | ValueCreatePopup | ProductAttributeValueModel model | Json |
| GET | /Admin/ProductAttribute/ValueEditPopup | ValueEditPopup | int id | View |
| POST | /Admin/ProductAttribute/ValueEditPopup | ValueEditPopup | ProductAttributeValueModel model | Json |

### Admin - SpecificationAttributeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/SpecificationAttribute | Index | - | Redirect to List |
| GET | /Admin/SpecificationAttribute/List | List | - | View |
| POST | /Admin/SpecificationAttribute/List | SpecificationAttributeList | SpecificationAttributeSearchModel searchModel | Json |
| GET | /Admin/SpecificationAttribute/Create | Create | - | View |
| POST | /Admin/SpecificationAttribute/Create | Create | SpecificationAttributeModel model, bool continueEditing | Redirect |
| GET | /Admin/SpecificationAttribute/Edit/{id} | Edit | int id | View |
| POST | /Admin/SpecificationAttribute/Edit | Edit | SpecificationAttributeModel model, bool continueEditing | Redirect |
| POST | /Admin/SpecificationAttribute/Delete | Delete | int id | Redirect |
| POST | /Admin/SpecificationAttribute/OptionList | OptionList | SpecificationAttributeOptionSearchModel searchModel | Json |
| POST | /Admin/SpecificationAttribute/OptionUpdate | OptionUpdate | SpecificationAttributeOptionModel model | NullJson |
| POST | /Admin/SpecificationAttribute/OptionDelete | OptionDelete | int id | NullJson |
| GET | /Admin/SpecificationAttribute/OptionCreatePopup | OptionCreatePopup | int specificationAttributeId | View |
| POST | /Admin/SpecificationAttribute/OptionCreatePopup | OptionCreatePopup | SpecificationAttributeOptionModel model | Json |
| GET | /Admin/SpecificationAttribute/OptionEditPopup | OptionEditPopup | int id | View |
| POST | /Admin/SpecificationAttribute/OptionEditPopup | OptionEditPopup | SpecificationAttributeOptionModel model | Json |

### Admin - FilterLevelValueController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/FilterLevelValue | Index | - | Redirect to List |
| GET | /Admin/FilterLevelValue/List | List | - | View |
| POST | /Admin/FilterLevelValue/List | FilterLevelValueList | FilterLevelValueSearchModel searchModel | Json |
| GET | /Admin/FilterLevelValue/Create | Create | - | View |
| POST | /Admin/FilterLevelValue/Create | Create | FilterLevelValueModel model, bool continueEditing | Redirect |
| GET | /Admin/FilterLevelValue/Edit/{id} | Edit | int id | View |
| POST | /Admin/FilterLevelValue/Edit | Edit | FilterLevelValueModel model, bool continueEditing | Redirect |
| POST | /Admin/FilterLevelValue/Delete | Delete | int id | Redirect |

---

## 2. 订单服务 (order)

### 前端 - OrderController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /order/history | CustomerOrders | int? pageNumber, OrderHistoryPeriods limit | View |
| GET | /order/recurringpayments | CustomerRecurringPayments | - | View |
| POST | /order/recurringpayments | CancelRecurringPayment | IFormCollection form | View |
| POST | /order/recurringpayments | RetryLastRecurringPayment | IFormCollection form | View |
| GET | /order/rewardpoints | CustomerRewardPoints | int? pageNumber | View |
| GET | /order/details/{orderId} | Details | int orderId | View |
| GET | /order/print/{orderId} | PrintOrderDetails | int orderId | View(Details) |
| GET | /order/pdf/{orderId} | GetPdfInvoice | int orderId | File(PDF) |
| POST | /order/cancel/{orderId} | CancelOrder | int orderId | Redirect |
| GET | /order/reorder/{orderId} | ReOrder | int orderId | Redirect |
| POST | /order/details | RePostPayment | int orderId | Redirect/EmptyResult |
| GET | /order/shipment/{shipmentId} | ShipmentDetails | int shipmentId | View |

### 前端 - ReturnRequestController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /returnrequest/history | CustomerReturnRequests | - | View |
| GET | /returnrequest/{orderId} | ReturnRequest | int orderId | View |
| POST | /returnrequest | ReturnRequestSubmit | ReturnRequestModel model | View/Redirect |
| POST | /returnrequest/uploadfile | UploadFileReturnRequest | - | Json |

### Admin - OrderController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Order | Index | - | Redirect to List |
| GET | /Admin/Order/List | List | - | View |
| POST | /Admin/Order/List | OrderList | OrderSearchModel searchModel | Json |
| POST | /Admin/Order/ReportAggregates | ReportAggregates | OrderSearchModel searchModel | Json |
| POST | /Admin/Order/GoToOrderId | GoToOrderId | int goToOrderId | Redirect |
| POST | /Admin/Order/ExportXmlAll | ExportXmlAll | OrderSearchModel searchModel | File(XML) |
| POST | /Admin/Order/ExportXmlSelected | ExportXmlSelected | ICollection<int> selectedIds | File(XML) |
| POST | /Admin/Order/ExportExcelAll | ExportExcelAll | OrderSearchModel searchModel | File(Excel) |
| POST | /Admin/Order/ExportExcelSelected | ExportExcelSelected | ICollection<int> selectedIds | File(Excel) |
| POST | /Admin/Order/ImportFromXlsx | ImportFromXlsx | IFormFile importexcelFile | Redirect |
| POST | /Admin/Order/CancelOrder | CancelOrder | int id | Redirect |
| POST | /Admin/Order/CaptureOrder | CaptureOrder | int id | Redirect |
| POST | /Admin/Order/MarkOrderAsPaid | MarkOrderAsPaid | int id | Redirect |
| POST | /Admin/Order/RefundOrder | RefundOrder | int id | Redirect |
| POST | /Admin/Order/RefundOrderOffline | RefundOrderOffline | int id | Redirect |
| POST | /Admin/Order/VoidOrder | VoidOrder | int id | Redirect |
| POST | /Admin/Order/VoidOrderOffline | VoidOrderOffline | int id | Redirect |
| GET | /Admin/Order/PartiallyRefundOrderPopup | PartiallyRefundOrderPopup | int id, bool online | View |
| POST | /Admin/Order/PartiallyRefundOrderPopup | PartiallyRefundOrderPopup | int id, bool online | Redirect |
| POST | /Admin/Order/ChangeOrderStatus | ChangeOrderStatus | int id, int orderStatusId | Redirect |
| GET | /Admin/Order/Edit/{id} | Edit | int id | View |
| POST | /Admin/Order/Delete | Delete | int id | Redirect |
| GET | /Admin/Order/PdfInvoice | PdfInvoice | int orderId | File(PDF) |
| POST | /Admin/Order/PdfInvoiceAll | PdfInvoiceAll | OrderSearchModel searchModel | File(ZIP) |
| POST | /Admin/Order/PdfInvoiceSelected | PdfInvoiceSelected | ICollection<int> selectedIds | File(ZIP) |
| POST | /Admin/Order/ProductDetails_AttributeChange | ProductDetails_AttributeChange | int productId, IFormCollection form | Json |
| POST | /Admin/Order/EditCreditCardInfo | EditCreditCardInfo | OrderModel model | Redirect |
| POST | /Admin/Order/EditOrderTotals | EditOrderTotals | OrderModel model | Redirect |
| POST | /Admin/Order/EditShippingMethod | EditShippingMethod | OrderModel model | Redirect |
| POST | /Admin/Order/EditOrderItem | EditOrderItem | int orderId, int orderItemId | Redirect |
| POST | /Admin/Order/DeleteOrderItem | DeleteOrderItem | int orderId, int orderItemId | Redirect |
| POST | /Admin/Order/ResetDownloadCount | ResetDownloadCount | int orderId, int orderItemId | Redirect |
| POST | /Admin/Order/ActivateDownloadItem | ActivateDownloadItem | int orderId, int orderItemId, bool activate | Redirect |
| GET | /Admin/Order/UploadLicenseFilePopup | UploadLicenseFilePopup | int orderId, int orderItemId | View |
| POST | /Admin/Order/UploadLicenseFilePopup | UploadLicenseFilePopup | OrderModel.UploadLicenseModel model | Redirect |
| POST | /Admin/Order/DeleteLicense | DeleteLicense | int orderId, int orderItemId | Redirect |
| GET | /Admin/Order/AddProductToOrder | AddProductToOrder | int orderId | View |
| POST | /Admin/Order/AddProductToOrder | AddProductToOrder | int orderId, AddProductToOrderSearchModel searchModel | Json |
| GET | /Admin/Order/AddProductToOrderDetails | AddProductToOrderDetails | int orderId, int productId | View |
| POST | /Admin/Order/AddProductToOrderDetails | AddProductToOrderDetails | int orderId, int productId, AddProductToOrderModel model | Redirect |
| GET | /Admin/Order/AddressEdit | AddressEdit | int addressId, int orderId | View |
| POST | /Admin/Order/AddressEdit | AddressEdit | OrderAddressModel model | Redirect |
| GET | /Admin/Order/ShipmentList | ShipmentList | - | View |
| POST | /Admin/Order/ShipmentListSelect | ShipmentListSelect | ShipmentSearchModel searchModel | Json |
| POST | /Admin/Order/ShipmentsByOrder | ShipmentsByOrder | ShipmentSearchModel searchModel | Json |
| POST | /Admin/Order/ShipmentsItemsByShipmentId | ShipmentsItemsByShipmentId | int shipmentId | Json |
| GET | /Admin/Order/AddShipment | AddShipment | int orderId | View |
| POST | /Admin/Order/AddShipment | AddShipment | ShipmentModel model | Redirect |
| GET | /Admin/Order/ShipmentDetails | ShipmentDetails | int id | View |
| POST | /Admin/Order/DeleteShipment | DeleteShipment | int id | Redirect |
| POST | /Admin/Order/SetTrackingNumber | SetTrackingNumber | ShipmentModel model | Redirect |
| POST | /Admin/Order/SetShipmentAdminComment | SetShipmentAdminComment | ShipmentModel model | Redirect |
| POST | /Admin/Order/ShipOrder | ShipOrder | int id | Redirect |
| POST | /Admin/Order/DeliverOrder | DeliverOrder | int id | Redirect |
| POST | /Admin/Order/OrderNoteList | OrderNoteList | OrderNoteSearchModel searchModel | Json |
| POST | /Admin/Order/OrderNoteAdd | OrderNoteAdd | int orderId, string note | Json |
| POST | /Admin/Order/OrderNoteDelete | OrderNoteDelete | int id | NullJson |

### Admin - RecurringPaymentController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/RecurringPayment | Index | - | Redirect to List |
| GET | /Admin/RecurringPayment/List | List | - | View |
| POST | /Admin/RecurringPayment/List | RecurringPaymentList | RecurringPaymentSearchModel searchModel | Json |
| GET | /Admin/RecurringPayment/Edit/{id} | Edit | int id | View |
| POST | /Admin/RecurringPayment/Delete | Delete | int id | Redirect |
| POST | /Admin/RecurringPayment/History | History | RecurringPaymentSearchModel searchModel | Json |

### Admin - GiftCardController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/GiftCard | Index | - | Redirect to List |
| GET | /Admin/GiftCard/List | List | - | View |
| POST | /Admin/GiftCard/List | GiftCardList | GiftCardSearchModel searchModel | Json |
| GET | /Admin/GiftCard/Create | Create | - | View |
| POST | /Admin/GiftCard/Create | Create | GiftCardModel model, bool continueEditing | Redirect |
| GET | /Admin/GiftCard/Edit/{id} | Edit | int id | View |
| POST | /Admin/GiftCard/Edit | Edit | GiftCardModel model, bool continueEditing | Redirect |
| POST | /Admin/GiftCard/Delete | Delete | int id | Redirect |
| POST | /Admin/GiftCard/GenerateCouponCode | GenerateCouponCode | - | Json |

### Admin - ReturnRequestController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/ReturnRequest | Index | - | Redirect to List |
| GET | /Admin/ReturnRequest/List | List | - | View |
| POST | /Admin/ReturnRequest/List | ReturnRequestList | ReturnRequestSearchModel searchModel | Json |
| GET | /Admin/ReturnRequest/Edit/{id} | Edit | int id | View |
| POST | /Admin/ReturnRequest/Edit | Edit | ReturnRequestModel model | Redirect |
| POST | /Admin/ReturnRequest/Delete | Delete | int id | Redirect |

---

## 3. 客户服务 (customer)

### 前端 - CustomerController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /login | Login | string returnUrl, bool? checkoutAsGuest | View |
| POST | /login | Login | LoginModel model, string returnUrl, bool? checkoutAsGuest | View/Redirect |
| GET | /logout | Logout | - | Redirect |
| GET | /register | Register | string returnUrl | View |
| POST | /register | Register | RegisterModel model, string returnUrl | View/Redirect |
| GET | /customer/info | Info | - | View |
| POST | /customer/info | Info | CustomerInfoModel model | View/Redirect |
| GET | /customer/addresses | Addresses | - | View |
| GET | /customer/addressadd | AddressAdd | - | View |
| POST | /customer/addressadd | AddressAdd | CustomerAddressEditModel model | Redirect |
| GET | /customer/addressedit/{addressId} | AddressEdit | int addressId | View |
| POST | /customer/addressedit | AddressEdit | CustomerAddressEditModel model | Redirect |
| GET | /customer/addressdelete/{addressId} | AddressDelete | int addressId | Redirect |
| GET | /customer/changepassword | ChangePassword | - | View |
| POST | /customer/changepassword | ChangePassword | ChangePasswordModel model | View |
| GET | /customer/avatar | Avatar | - | View |
| POST | /customer/avatar | Avatar | CustomerAvatarModel model | View |
| GET | /customer/gdprtools | GdprTools | - | View |
| POST | /customer/gdprtools | GdprTools | CustomerGdprModel model | View/Redirect |
| GET | /customer/checkusernameavailability | CheckUsernameAvailability | string username | Json |
| GET | /customer/activation | AccountActivation | string token, string email, string username | View |
| GET | /customer/emailrevalidation | EmailRevalidation | string token, string email | View |
| POST | /customer/deletedpmmrequest | DeleteDpmRequest | - | Redirect |
| POST | /customer/exportdpmmrequest | ExportDpmRequest | - | Redirect/File |

### 前端 - ProfileController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /profile/{id} | Index | int id, int? pageNumber | View |

### Admin - CustomerController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Customer | Index | - | Redirect to List |
| GET | /Admin/Customer/List | List | - | View |
| POST | /Admin/Customer/List | CustomerList | CustomerSearchModel searchModel | Json |
| GET | /Admin/Customer/Create | Create | - | View |
| POST | /Admin/Customer/Create | Create | CustomerModel model, bool continueEditing | Redirect |
| GET | /Admin/Customer/Edit/{id} | Edit | int id | View |
| POST | /Admin/Customer/Edit | Edit | CustomerModel model, bool continueEditing | Redirect |
| POST | /Admin/Customer/ChangePassword | ChangePassword | int id, string password | Redirect |
| POST | /Admin/Customer/MarkVatNumberAsValid | MarkVatNumberAsValid | int id | Redirect |
| POST | /Admin/Customer/MarkVatNumberAsInvalid | MarkVatNumberAsInvalid | int id | Redirect |
| POST | /Admin/Customer/RemoveAffiliate | RemoveAffiliate | int id | Redirect |
| POST | /Admin/Customer/RemoveBindMFA | RemoveBindMFA | int id | Redirect |
| POST | /Admin/Customer/Delete | Delete | int id | Redirect |
| POST | /Admin/Customer/Impersonate | Impersonate | int id | Redirect |
| POST | /Admin/Customer/SendWelcomeMessage | SendWelcomeMessage | int id | Redirect |
| POST | /Admin/Customer/ReSendActivationMessage | ReSendActivationMessage | int id | Redirect |
| POST | /Admin/Customer/SendEmail | SendEmail | int id, string subject, string body | Redirect |
| POST | /Admin/Customer/SendPm | SendPm | int id, string subject, string body | Redirect |
| POST | /Admin/Customer/RewardPointsHistorySelect | RewardPointsHistorySelect | RewardPointsSearchModel searchModel | Json |
| POST | /Admin/Customer/RewardPointsHistoryAdd | RewardPointsHistoryAdd | int customerId, int addValue, string message, bool activate | Json |
| POST | /Admin/Customer/AddressesSelect | AddressesSelect | CustomerAddressSearchModel searchModel | Json |
| POST | /Admin/Customer/AddressDelete | AddressDelete | int id | NullJson |
| GET | /Admin/Customer/AddressCreate | AddressCreate | int customerId | View |
| POST | /Admin/Customer/AddressCreate | AddressCreate | CustomerAddressModel model | Redirect |
| GET | /Admin/Customer/AddressEdit | AddressEdit | int id, int customerId | View |
| POST | /Admin/Customer/AddressEdit | AddressEdit | CustomerAddressModel model | Redirect |
| POST | /Admin/Customer/OrderList | OrderList | CustomerOrderSearchModel searchModel | Json |
| POST | /Admin/Customer/LoadCustomerStatistics | LoadCustomerStatistics | string period | Json |
| POST | /Admin/Customer/GetCartList | GetCartList | CustomerShoppingCartSearchModel searchModel | Json |
| POST | /Admin/Customer/ListActivityLog | ListActivityLog | CustomerActivityLogSearchModel searchModel | Json |
| POST | /Admin/Customer/BackInStockSubscriptionList | BackInStockSubscriptionList | CustomerBackInStockSubscriptionSearchModel searchModel | Json |
| GET | /Admin/Customer/GdprLog | GdprLog | int id | View |
| POST | /Admin/Customer/GdprLogList | GdprLogList | CustomerGdprLogSearchModel searchModel | Json |
| POST | /Admin/Customer/GdprDelete | GdprDelete | int id | Redirect |
| GET | /Admin/Customer/GdprExport | GdprExport | int id | File(Excel) |
| POST | /Admin/Customer/ExportExcelAll | ExportExcelAll | CustomerSearchModel searchModel | File(Excel) |
| POST | /Admin/Customer/ExportExcelSelected | ExportExcelSelected | ICollection<int> selectedIds | File(Excel) |
| POST | /Admin/Customer/ExportXmlAll | ExportXmlAll | CustomerSearchModel searchModel | File(XML) |
| POST | /Admin/Customer/ExportXmlSelected | ExportXmlSelected | ICollection<int> selectedIds | File(XML) |
| POST | /Admin/Customer/ImportExcel | ImportExcel | IFormFile importexcelFile | Redirect |

### Admin - CustomerRoleController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/CustomerRole | Index | - | Redirect to List |
| GET | /Admin/CustomerRole/List | List | - | View |
| POST | /Admin/CustomerRole/List | CustomerRoleList | CustomerRoleSearchModel searchModel | Json |
| GET | /Admin/CustomerRole/Create | Create | - | View |
| POST | /Admin/CustomerRole/Create | Create | CustomerRoleModel model, bool continueEditing | Redirect |
| GET | /Admin/CustomerRole/Edit/{id} | Edit | int id | View |
| POST | /Admin/CustomerRole/Edit | Edit | CustomerRoleModel model, bool continueEditing | Redirect |
| POST | /Admin/CustomerRole/Delete | Delete | int id | Redirect |

### Admin - CustomerAttributeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/CustomerAttribute | Index | - | Redirect to List |
| GET | /Admin/CustomerAttribute/List | List | - | View |
| POST | /Admin/CustomerAttribute/List | CustomerAttributeList | CustomerAttributeSearchModel searchModel | Json |
| GET | /Admin/CustomerAttribute/Create | Create | - | View |
| POST | /Admin/CustomerAttribute/Create | Create | CustomerAttributeModel model, bool continueEditing | Redirect |
| GET | /Admin/CustomerAttribute/Edit/{id} | Edit | int id | View |
| POST | /Admin/CustomerAttribute/Edit | Edit | CustomerAttributeModel model, bool continueEditing | Redirect |
| POST | /Admin/CustomerAttribute/Delete | Delete | int id | Redirect |
| POST | /Admin/CustomerAttribute/ValueList | ValueList | CustomerAttributeValueSearchModel searchModel | Json |
| POST | /Admin/CustomerAttribute/ValueUpdate | ValueUpdate | CustomerAttributeValueModel model | NullJson |
| POST | /Admin/CustomerAttribute/ValueDelete | ValueDelete | int id | NullJson |
| GET | /Admin/CustomerAttribute/ValueCreatePopup | ValueCreatePopup | int customerAttributeId | View |
| POST | /Admin/CustomerAttribute/ValueCreatePopup | ValueCreatePopup | CustomerAttributeValueModel model | Json |
| GET | /Admin/CustomerAttribute/ValueEditPopup | ValueEditPopup | int id | View |
| POST | /Admin/CustomerAttribute/ValueEditPopup | ValueEditPopup | CustomerAttributeValueModel model | Json |

### Admin - OnlineCustomerController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/OnlineCustomer | Index | - | Redirect to List |
| GET | /Admin/OnlineCustomer/List | List | - | View |
| POST | /Admin/OnlineCustomer/List | OnlineCustomerList | OnlineCustomerSearchModel searchModel | Json |

---

## 4. 支付服务 (payment)

### 前端 - CheckoutController (支付相关部分)
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /checkout/paymentmethod | PaymentMethod | - | View |
| POST | /checkout/paymentmethod | PaymentMethod | CheckoutPaymentMethodModel model | View/Redirect |
| GET | /checkout/paymentinfo | PaymentInfo | - | View |
| POST | /checkout/paymentinfo | PaymentInfo | - | View/Redirect |
| GET | /checkout/confirm | Confirm | - | View |
| POST | /checkout/confirm | Confirm | - | View/Redirect |
| POST | /checkout/opcsavepaymentmethod | OpcSavePaymentMethod | string paymentmethod | Json |
| POST | /checkout/opcsavepaymentinfo | OpcSavePaymentInfo | - | Json |

### Admin - PaymentController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Payment/Methods | PaymentMethods | - | Redirect |
| GET | /Admin/Payment/Methods | Methods | - | View |
| POST | /Admin/Payment/Methods | Methods | PaymentMethodListModel model | View |
| POST | /Admin/Payment/MethodUpdate | MethodUpdate | PaymentMethodModel model | NullJson |
| GET | /Admin/Payment/MethodRestrictions | MethodRestrictions | - | View |
| POST | /Admin/Payment/MethodRestrictions | MethodRestrictions | PaymentMethodRestrictionModel model | View |

---

## 5. 配送服务 (shipping)

### 前端 - CheckoutController (配送相关部分)
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /checkout/shippingaddress | ShippingAddress | - | View |
| POST | /checkout/shippingaddress | ShippingAddress | CheckoutShippingAddressModel model | View/Redirect |
| GET | /checkout/selectshippingaddress | SelectShippingAddress | int addressId | Redirect |
| GET | /checkout/shippingmethod | ShippingMethod | - | View |
| POST | /checkout/shippingmethod | ShippingMethod | CheckoutShippingMethodModel model | View/Redirect |
| POST | /checkout/opcsaveshipping | OpcSaveShipping | int shipping_option | Json |
| POST | /checkout/opcsaveshippingmethod | OpcSaveShippingMethod | string shippingoption | Json |

### Admin - ShippingController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Shipping/Providers | Providers | - | View |
| POST | /Admin/Shipping/Providers | Providers | ShippingProviderSearchModel searchModel | Json |
| POST | /Admin/Shipping/ProviderUpdate | ProviderUpdate | ShippingProviderModel model | NullJson |
| GET | /Admin/Shipping/PickupPointProviders | PickupPointProviders | - | View |
| POST | /Admin/Shipping/PickupPointProviders | PickupPointProviders | PickupPointProviderSearchModel searchModel | Json |
| POST | /Admin/Shipping/PickupPointProviderUpdate | PickupPointProviderUpdate | PickupPointProviderModel model | NullJson |
| GET | /Admin/Shipping/Methods | Methods | - | View |
| POST | /Admin/Shipping/Methods | Methods | ShippingMethodSearchModel searchModel | Json |
| GET | /Admin/Shipping/CreateMethod | CreateMethod | - | View |
| POST | /Admin/Shipping/CreateMethod | CreateMethod | ShippingMethodModel model, bool continueEditing | Redirect |
| GET | /Admin/Shipping/EditMethod/{id} | EditMethod | int id | View |
| POST | /Admin/Shipping/EditMethod | EditMethod | ShippingMethodModel model, bool continueEditing | Redirect |
| POST | /Admin/Shipping/DeleteMethod | DeleteMethod | int id | Redirect |
| GET | /Admin/Shipping/DatesAndRanges | DatesAndRanges | - | View |
| POST | /Admin/Shipping/DeliveryDates | DeliveryDates | DeliveryDateSearchModel searchModel | Json |
| GET | /Admin/Shipping/CreateDeliveryDate | CreateDeliveryDate | - | View |
| POST | /Admin/Shipping/CreateDeliveryDate | CreateDeliveryDate | DeliveryDateModel model, bool continueEditing | Redirect |
| GET | /Admin/Shipping/EditDeliveryDate/{id} | EditDeliveryDate | int id | View |
| POST | /Admin/Shipping/EditDeliveryDate | EditDeliveryDate | DeliveryDateModel model, bool continueEditing | Redirect |
| POST | /Admin/Shipping/DeleteDeliveryDate | DeleteDeliveryDate | int id | Redirect |
| POST | /Admin/Shipping/ProductAvailabilityRanges | ProductAvailabilityRanges | ProductAvailabilityRangeSearchModel searchModel | Json |
| GET | /Admin/Shipping/CreateProductAvailabilityRange | CreateProductAvailabilityRange | - | View |
| POST | /Admin/Shipping/CreateProductAvailabilityRange | CreateProductAvailabilityRange | ProductAvailabilityRangeModel model, bool continueEditing | Redirect |
| GET | /Admin/Shipping/EditProductAvailabilityRange/{id} | EditProductAvailabilityRange | int id | View |
| POST | /Admin/Shipping/EditProductAvailabilityRange | EditProductAvailabilityRange | ProductAvailabilityRangeModel model, bool continueEditing | Redirect |
| POST | /Admin/Shipping/DeleteProductAvailabilityRange | DeleteProductAvailabilityRange | int id | Redirect |
| POST | /Admin/Shipping/Warehouses | Warehouses | WarehouseSearchModel searchModel | Json |
| GET | /Admin/Shipping/CreateWarehouse | CreateWarehouse | - | View |
| POST | /Admin/Shipping/CreateWarehouse | CreateWarehouse | WarehouseModel model, bool continueEditing | Redirect |
| GET | /Admin/Shipping/EditWarehouse/{id} | EditWarehouse | int id | View |
| POST | /Admin/Shipping/EditWarehouse | EditWarehouse | WarehouseModel model, bool continueEditing | Redirect |
| POST | /Admin/Shipping/DeleteWarehouse | DeleteWarehouse | int id | Redirect |
| GET | /Admin/Shipping/Restrictions | Restrictions | - | View |
| POST | /Admin/Shipping/Restrictions | Restrictions | ShippingRestrictionModel model | View |

---

## 6. 折扣服务 (discount)

### Admin - DiscountController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Discount | Index | - | Redirect to List |
| GET | /Admin/Discount/List | List | - | View |
| POST | /Admin/Discount/List | DiscountList | DiscountSearchModel searchModel | Json |
| GET | /Admin/Discount/Create | Create | - | View |
| POST | /Admin/Discount/Create | Create | DiscountModel model, bool continueEditing | Redirect |
| GET | /Admin/Discount/Edit/{id} | Edit | int id | View |
| POST | /Admin/Discount/Edit | Edit | DiscountModel model, bool continueEditing | Redirect |
| POST | /Admin/Discount/Delete | Delete | int id | Redirect |
| GET | /Admin/Discount/GetDiscountRequirementConfigurationUrl | GetDiscountRequirementConfigurationUrl | int discountRequirementId, int discountId | Json |
| GET | /Admin/Discount/GetDiscountRequirements | GetDiscountRequirements | int discountId | Json |
| POST | /Admin/Discount/AddNewGroup | AddNewGroup | int discountId | Json |
| GET | /Admin/Discount/CouponCodeReservedWarning | CouponCodeReservedWarning | string couponCode, int discountId | Json |
| POST | /Admin/Discount/ProductList | ProductList | DiscountProductSearchModel searchModel | Json |
| POST | /Admin/Discount/ProductDelete | ProductDelete | int id | NullJson |
| GET | /Admin/Discount/ProductAddPopup | ProductAddPopup | int discountId | View |
| POST | /Admin/Discount/ProductAddPopup | ProductAddPopup | - | Json(list) |
| POST | /Admin/Discount/CategoryList | CategoryList | DiscountCategorySearchModel searchModel | Json |
| POST | /Admin/Discount/CategoryDelete | CategoryDelete | int id | NullJson |
| GET | /Admin/Discount/CategoryAddPopup | CategoryAddPopup | int discountId | View |
| POST | /Admin/Discount/CategoryAddPopup | CategoryAddPopup | - | Json(list) |
| POST | /Admin/Discount/ManufacturerList | ManufacturerList | DiscountManufacturerSearchModel searchModel | Json |
| POST | /Admin/Discount/ManufacturerDelete | ManufacturerDelete | int id | NullJson |
| GET | /Admin/Discount/ManufacturerAddPopup | ManufacturerAddPopup | int discountId | View |
| POST | /Admin/Discount/ManufacturerAddPopup | ManufacturerAddPopup | - | Json(list) |
| POST | /Admin/Discount/UsageHistoryList | UsageHistoryList | DiscountUsageHistorySearchModel searchModel | Json |
| POST | /Admin/Discount/UsageHistoryDelete | UsageHistoryDelete | int id | NullJson |

---

## 7. 税务服务 (tax)

### Admin - TaxController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Tax | List | - | Redirect |
| GET | /Admin/Tax/Providers | Providers | - | View |
| POST | /Admin/Tax/Providers | Providers | TaxProviderSearchModel searchModel | Json |
| GET | /Admin/Tax/MarkAsPrimaryProvider | MarkAsPrimaryProvider | int id | Redirect |
| GET | /Admin/Tax/Categories | Categories | - | View |
| POST | /Admin/Tax/Categories | Categories | TaxCategorySearchModel searchModel | Json |
| POST | /Admin/Tax/CategoryUpdate | CategoryUpdate | TaxCategoryModel model | NullJson |
| POST | /Admin/Tax/CategoryAdd | CategoryAdd | TaxCategoryModel model | Json |
| POST | /Admin/Tax/CategoryDelete | CategoryDelete | int id | NullJson |

---

## 8. 供应商服务 (vendorsvc)

### 前端 - VendorController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /vendor/apply | ApplyVendor | - | View |
| POST | /vendor/apply | ApplyVendor | VendorInfoModel model | View/Redirect |
| GET | /vendor/info | Info | - | View |
| POST | /vendor/info | Info | VendorInfoModel model | View/Redirect |
| POST | /vendor/removepicture | RemovePicture | - | Redirect |

### Admin - VendorController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Vendor | Index | - | Redirect to List |
| GET | /Admin/Vendor/List | List | - | View |
| POST | /Admin/Vendor/List | VendorList | VendorSearchModel searchModel | Json |
| GET | /Admin/Vendor/Create | Create | - | View |
| POST | /Admin/Vendor/Create | Create | VendorModel model, bool continueEditing | Redirect |
| GET | /Admin/Vendor/Edit/{id} | Edit | int id | View |
| POST | /Admin/Vendor/Edit | Edit | VendorModel model, bool continueEditing | Redirect |
| POST | /Admin/Vendor/Delete | Delete | int id | Redirect |
| POST | /Admin/Vendor/VendorNotesSelect | VendorNotesSelect | VendorNoteSearchModel searchModel | Json |
| POST | /Admin/Vendor/VendorNoteAdd | VendorNoteAdd | int vendorId, string message | Json |
| POST | /Admin/Vendor/VendorNoteDelete | VendorNoteDelete | int id | NullJson |
| GET | /Admin/Vendor/AddCustomerToVendorPopup | AddCustomerToVendorPopup | int vendorId | View |
| POST | /Admin/Vendor/AddCustomerToVendorPopup | AddCustomerToVendorPopup | - | Json(list) |

### Admin - VendorAttributeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/VendorAttribute | Index | - | Redirect to List |
| GET | /Admin/VendorAttribute/List | List | - | View |
| POST | /Admin/VendorAttribute/List | VendorAttributeList | VendorAttributeSearchModel searchModel | Json |
| GET | /Admin/VendorAttribute/Create | Create | - | View |
| POST | /Admin/VendorAttribute/Create | Create | VendorAttributeModel model, bool continueEditing | Redirect |
| GET | /Admin/VendorAttribute/Edit/{id} | Edit | int id | View |
| POST | /Admin/VendorAttribute/Edit | Edit | VendorAttributeModel model, bool continueEditing | Redirect |
| POST | /Admin/VendorAttribute/Delete | Delete | int id | Redirect |
| POST | /Admin/VendorAttribute/ValueList | ValueList | VendorAttributeValueSearchModel searchModel | Json |
| POST | /Admin/VendorAttribute/ValueUpdate | ValueUpdate | VendorAttributeValueModel model | NullJson |
| POST | /Admin/VendorAttribute/ValueDelete | ValueDelete | int id | NullJson |
| GET | /Admin/VendorAttribute/ValueCreatePopup | ValueCreatePopup | int vendorAttributeId | View |
| POST | /Admin/VendorAttribute/ValueCreatePopup | ValueCreatePopup | VendorAttributeValueModel model | Json |
| GET | /Admin/VendorAttribute/ValueEditPopup | ValueEditPopup | int id | View |
| POST | /Admin/VendorAttribute/ValueEditPopup | ValueEditPopup | VendorAttributeValueModel model | Json |

---

## 9. 内容服务 (content)

### 前端 - BlogController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /blog | List | BlogPagingFilteringModel command | View |
| GET | /blog/tag/{tag} | BlogByTag | string tag, BlogPagingFilteringModel command | View |
| GET | /blog/month/{month} | BlogByMonth | string month, BlogPagingFilteringModel command | View |
| GET | /blog/rss | ListRss | - | RssActionResult |
| GET | /blog/{blogPostId} | BlogPost | int blogPostId | View |
| POST | /blog/{blogPostId} | BlogCommentAdd | BlogPostModel model | View/Redirect |

### 前端 - NewsController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /news | List | NewsPagingFilteringModel command | View |
| GET | /news/rss | ListRss | - | RssActionResult |
| GET | /news/{newsItemId} | NewsItem | int newsItemId | View |
| POST | /news/{newsItemId} | NewsCommentAdd | NewsItemModel model | View/Redirect |

### 前端 - TopicController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /t/{systemName} | TopicDetails | string systemName | View |
| GET | /topic/popup/{systemName} | TopicDetailsPopup | string systemName | PartialView |
| POST | /topic/authenticate | Authenticate | int id, string password | Json |

### 前端 - PollController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| POST | /poll/vote | Vote | int pollAnswerId | Json |

### 前端 - BoardsController (论坛)
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /boards | Index | - | View |
| GET | /boards/activediscussions | ActiveDiscussions | int? forumId, int? pageNumber | View |
| GET | /boards/activediscussionsrss | ActiveDiscussionsRss | int? forumId | RssActionResult |
| GET | /boards/forumgroup/{id} | ForumGroup | int id | View |
| GET | /boards/forum/{id} | Forum | int id, int? pageNumber | View |
| GET | /boards/topic/{id} | Topic | int id, int? pageNumber | View |
| GET | /boards/topic/create/{id} | TopicCreate | int id | View |
| POST | /boards/topic/create | TopicCreate | ForumTopicCreateModel model | Redirect |
| GET | /boards/topic/edit/{id} | TopicEdit | int id | View |
| POST | /boards/topic/edit | TopicEdit | ForumTopicEditModel model | Redirect |
| POST | /boards/topic/delete | TopicDelete | int id | Redirect |
| GET | /boards/post/create/{id} | PostCreate | int id, int? quote | View |
| POST | /boards/post/create | PostCreate | ForumPostCreateModel model | Redirect |
| GET | /boards/post/edit/{id} | PostEdit | int id | View |
| POST | /boards/post/edit | PostEdit | ForumPostEditModel model | Redirect |
| POST | /boards/post/delete | PostDelete | int id | Redirect |
| GET | /boards/search | Search | string searchterms, int? forumId | View |
| POST | /boards/subscribe/{id} | Subscribe | int id | Redirect |
| POST | /boards/unsubscribe/{id} | Unsubscribe | int id | Redirect |
| POST | /boards/postvote | PostVote | int postId, bool isUp | Json |

### Admin - BlogController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Blog | Index | - | Redirect |
| GET | /Admin/Blog/List | BlogPosts | - | View |
| POST | /Admin/Blog/List | List | BlogPostSearchModel searchModel | Json |
| GET | /Admin/Blog/Create | BlogPostCreate | - | View |
| POST | /Admin/Blog/Create | BlogPostCreate | BlogPostModel model, bool continueEditing | Redirect |
| GET | /Admin/Blog/Edit/{id} | BlogPostEdit | int id | View |
| POST | /Admin/Blog/Edit | BlogPostEdit | BlogPostModel model, bool continueEditing | Redirect |
| POST | /Admin/Blog/Delete | Delete | int id | Redirect |
| GET | /Admin/Blog/Comments | BlogComments | - | View |
| POST | /Admin/Blog/Comments | Comments | BlogCommentSearchModel searchModel | Json |
| POST | /Admin/Blog/CommentUpdate | CommentUpdate | BlogCommentModel model | NullJson |
| POST | /Admin/Blog/CommentDelete | CommentDelete | int id | NullJson |
| POST | /Admin/Blog/DeleteSelectedComments | DeleteSelectedComments | ICollection<int> selectedIds | Json |
| POST | /Admin/Blog/ApproveSelected | ApproveSelected | ICollection<int> selectedIds | Json |
| POST | /Admin/Blog/DisapproveSelected | DisapproveSelected | ICollection<int> selectedIds | Json |

### Admin - NewsController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/News | Index | - | Redirect |
| GET | /Admin/News/List | NewsItems | - | View |
| POST | /Admin/News/List | List | NewsItemSearchModel searchModel | Json |
| GET | /Admin/News/Create | NewsItemCreate | - | View |
| POST | /Admin/News/Create | NewsItemCreate | NewsItemModel model, bool continueEditing | Redirect |
| GET | /Admin/News/Edit/{id} | NewsItemEdit | int id | View |
| POST | /Admin/News/Edit | NewsItemEdit | NewsItemModel model, bool continueEditing | Redirect |
| POST | /Admin/News/Delete | Delete | int id | Redirect |
| GET | /Admin/News/Comments | NewsComments | - | View |
| POST | /Admin/News/Comments | Comments | NewsCommentSearchModel searchModel | Json |
| POST | /Admin/News/CommentUpdate | CommentUpdate | NewsCommentModel model | NullJson |
| POST | /Admin/News/CommentDelete | CommentDelete | int id | NullJson |
| POST | /Admin/News/DeleteSelectedComments | DeleteSelectedComments | ICollection<int> selectedIds | Json |
| POST | /Admin/News/ApproveSelected | ApproveSelected | ICollection<int> selectedIds | Json |
| POST | /Admin/News/DisapproveSelected | DisapproveSelected | ICollection<int> selectedIds | Json |

### Admin - ForumController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Forum | Index | - | Redirect to List |
| GET | /Admin/Forum/List | List | - | View |
| POST | /Admin/Forum/List | ForumList | ForumSearchModel searchModel | Json |
| GET | /Admin/Forum/Create | Create | - | View |
| POST | /Admin/Forum/Create | Create | ForumModel model, bool continueEditing | Redirect |
| GET | /Admin/Forum/Edit/{id} | Edit | int id | View |
| POST | /Admin/Forum/Edit | Edit | ForumModel model, bool continueEditing | Redirect |
| POST | /Admin/Forum/Delete | Delete | int id | Redirect |
| POST | /Admin/Forum/ForumGroupList | ForumGroupList | ForumGroupSearchModel searchModel | Json |
| GET | /Admin/Forum/CreateForumGroup | CreateForumGroup | - | View |
| POST | /Admin/Forum/CreateForumGroup | CreateForumGroup | ForumGroupModel model, bool continueEditing | Redirect |
| GET | /Admin/Forum/EditForumGroup/{id} | EditForumGroup | int id | View |
| POST | /Admin/Forum/EditForumGroup | EditForumGroup | ForumGroupModel model, bool continueEditing | Redirect |
| POST | /Admin/Forum/DeleteForumGroup | DeleteForumGroup | int id | Redirect |

### Admin - TopicController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Topic | Index | - | Redirect to List |
| GET | /Admin/Topic/List | List | - | View |
| POST | /Admin/Topic/List | TopicList | TopicSearchModel searchModel | Json |
| GET | /Admin/Topic/Create | Create | - | View |
| POST | /Admin/Topic/Create | Create | TopicModel model, bool continueEditing | Redirect |
| GET | /Admin/Topic/Edit/{id} | Edit | int id | View |
| POST | /Admin/Topic/Edit | Edit | TopicModel model, bool continueEditing | Redirect |
| POST | /Admin/Topic/Delete | Delete | int id | Redirect |

### Admin - PollController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Poll | Index | - | Redirect to List |
| GET | /Admin/Poll/List | List | - | View |
| POST | /Admin/Poll/List | PollList | PollSearchModel searchModel | Json |
| GET | /Admin/Poll/Create | Create | - | View |
| POST | /Admin/Poll/Create | Create | PollModel model, bool continueEditing | Redirect |
| GET | /Admin/Poll/Edit/{id} | Edit | int id | View |
| POST | /Admin/Poll/Edit | Edit | PollModel model, bool continueEditing | Redirect |
| POST | /Admin/Poll/Delete | Delete | int id | Redirect |
| POST | /Admin/Poll/AnswerList | AnswerList | PollAnswerSearchModel searchModel | Json |
| POST | /Admin/Poll/AnswerUpdate | AnswerUpdate | PollAnswerModel model | NullJson |
| POST | /Admin/Poll/AnswerDelete | AnswerDelete | int id | NullJson |

### Admin - ReviewTypeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/ReviewType | Index | - | Redirect to List |
| GET | /Admin/ReviewType/List | List | - | View |
| POST | /Admin/ReviewType/List | ReviewTypeList | ReviewTypeSearchModel searchModel | Json |
| GET | /Admin/ReviewType/Create | Create | - | View |
| POST | /Admin/ReviewType/Create | Create | ReviewTypeModel model, bool continueEditing | Redirect |
| GET | /Admin/ReviewType/Edit/{id} | Edit | int id | View |
| POST | /Admin/ReviewType/Edit | Edit | ReviewTypeModel model, bool continueEditing | Redirect |
| POST | /Admin/ReviewType/Delete | Delete | int id | Redirect |

---

## 10. 消息服务 (message)

### 前端 - NewsletterController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| POST | /subscribenewsletter | SubscribeNewsletter | string email, bool subscribe | Json |
| GET | /newsletter/subscriptionactivation | SubscriptionActivation | string token, bool active | View/Redirect |

### 前端 - PrivateMessagesController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /privatemessages/{tab?} | Index | int? tab, int? pageNumber | View |
| POST | /privatemessages/deleteinboxpm | DeleteInboxPM | IFormCollection form | Redirect |
| POST | /privatemessages/markunread | MarkUnread | IFormCollection form | Redirect |
| POST | /privatemessages/deletesentpm | DeleteSentPM | IFormCollection form | Redirect |
| GET | /privatemessages/send/{to?} | SendPM | int to | View |
| POST | /privatemessages/send | SendPM | PrivateMessageModel model | Redirect |
| GET | /privatemessages/view/{privateMessageId} | ViewPM | int privateMessageId | View |
| GET | /privatemessages/delete/{privateMessageId} | DeletePM | int privateMessageId | Redirect |

### Admin - EmailAccountController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/EmailAccount | Index | - | Redirect to List |
| GET | /Admin/EmailAccount/List | List | - | View |
| POST | /Admin/EmailAccount/List | EmailAccountList | EmailAccountSearchModel searchModel | Json |
| GET | /Admin/EmailAccount/Create | Create | - | View |
| POST | /Admin/EmailAccount/Create | Create | EmailAccountModel model, bool continueEditing | Redirect |
| GET | /Admin/EmailAccount/Edit/{id} | Edit | int id | View |
| POST | /Admin/EmailAccount/Edit | Edit | EmailAccountModel model, bool continueEditing | Redirect |
| POST | /Admin/EmailAccount/Delete | Delete | int id | Redirect |

### Admin - MessageTemplateController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/MessageTemplate | Index | - | Redirect to List |
| GET | /Admin/MessageTemplate/List | List | - | View |
| POST | /Admin/MessageTemplate/List | MessageTemplateList | MessageTemplateSearchModel searchModel | Json |
| GET | /Admin/MessageTemplate/Edit/{id} | Edit | int id | View |
| POST | /Admin/MessageTemplate/Edit | Edit | MessageTemplateModel model, bool continueEditing | Redirect |
| POST | /Admin/MessageTemplate/Delete | Delete | int id | Redirect |
| GET | /Admin/MessageTemplate/TestTemplate | TestTemplate | int id | View |
| POST | /Admin/MessageTemplate/TestTemplate | TestTemplate | int id | View |

### Admin - CampaignController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Campaign | Index | - | Redirect to List |
| GET | /Admin/Campaign/List | List | - | View |
| POST | /Admin/Campaign/List | CampaignList | CampaignSearchModel searchModel | Json |
| GET | /Admin/Campaign/Create | Create | - | View |
| POST | /Admin/Campaign/Create | Create | CampaignModel model, bool continueEditing | Redirect |
| GET | /Admin/Campaign/Edit/{id} | Edit | int id | View |
| POST | /Admin/Campaign/Edit | Edit | CampaignModel model, bool continueEditing | Redirect |
| POST | /Admin/Campaign/Delete | Delete | int id | Redirect |
| GET | /Admin/Campaign/SendTestEmail | SendTestEmail | int id | Redirect |

### Admin - QueuedEmailController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/QueuedEmail | Index | - | Redirect to List |
| GET | /Admin/QueuedEmail/List | List | - | View |
| POST | /Admin/QueuedEmail/List | QueuedEmailList | QueuedEmailSearchModel searchModel | Json |
| GET | /Admin/QueuedEmail/Edit/{id} | Edit | int id | View |
| POST | /Admin/QueuedEmail/Edit | Edit | QueuedEmailModel model, bool continueEditing | Redirect |
| POST | /Admin/QueuedEmail/Delete | Delete | int id | Redirect |
| POST | /Admin/QueuedEmail/Requeue | Requeue | int id | Redirect |
| POST | /Admin/QueuedEmail/DeleteAll | DeleteAll | - | Redirect |

### Admin - NewsLetterSubscriptionController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/NewsLetterSubscription | Index | - | Redirect to List |
| GET | /Admin/NewsLetterSubscription/List | List | - | View |
| POST | /Admin/NewsLetterSubscription/List | SubscriptionList | NewsLetterSubscriptionSearchModel searchModel | Json |
| POST | /Admin/NewsLetterSubscription/ExportCsv | ExportCsv | NewsLetterSubscriptionSearchModel searchModel | File(CSV) |
| POST | /Admin/NewsLetterSubscription/ExportExcel | ExportExcel | NewsLetterSubscriptionSearchModel searchModel | File(Excel) |
| POST | /Admin/NewsLetterSubscription/ImportCsv | ImportCsv | IFormFile importcsvfile | Redirect |
| POST | /Admin/NewsLetterSubscription/ImportExcel | ImportExcel | IFormFile importexcelfile | Redirect |

### Admin - NewsLetterSubscriptionTypeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/NewsLetterSubscriptionType | Index | - | Redirect to List |
| GET | /Admin/NewsLetterSubscriptionType/List | List | - | View |
| POST | /Admin/NewsLetterSubscriptionType/List | SubscriptionTypeList | NewsLetterSubscriptionTypeSearchModel searchModel | Json |
| GET | /Admin/NewsLetterSubscriptionType/Create | Create | - | View |
| POST | /Admin/NewsLetterSubscriptionType/Create | Create | NewsLetterSubscriptionTypeModel model, bool continueEditing | Redirect |
| GET | /Admin/NewsLetterSubscriptionType/Edit/{id} | Edit | int id | View |
| POST | /Admin/NewsLetterSubscriptionType/Edit | Edit | NewsLetterSubscriptionTypeModel model, bool continueEditing | Redirect |
| POST | /Admin/NewsLetterSubscriptionType/Delete | Delete | int id | Redirect |

---

## 11. 媒体服务 (media)

### 前端 - DownloadController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /download/sample/{productId} | Sample | int productId | File/Redirect |
| GET | /download/getdownload/{orderItemId} | GetDownload | int orderItemId, bool agree | File/Redirect |
| GET | /download/getlicense/{orderItemId} | GetLicense | int orderItemId | File/Redirect |
| GET | /download/getfileupload | GetFileUpload | int downloadId | File/Redirect |
| GET | /download/getordernotefile/{orderNoteId} | GetOrderNoteFile | int orderNoteId | File/Redirect |

### Admin - PictureController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Picture | Index | - | Redirect to List |
| GET | /Admin/Picture/List | List | - | View |
| POST | /Admin/Picture/List | PictureList | PictureSearchModel searchModel | Json |
| POST | /Admin/Picture/Delete | Delete | int id | Redirect |
| POST | /Admin/Picture/PictureDelete | PictureDelete | int id | NullJson |

### Admin - DownloadController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Download/DownloadFile | DownloadFile | int downloadId | File/Redirect |
| POST | /Admin/Download/SaveFile | SaveFile | IFormFile downloadFile | Json |
| POST | /Admin/Download/DeleteFile | DeleteFile | int downloadId | Json |

### Admin - RoxyFilemanController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/RoxyFileman/ProcessRequest | ProcessRequest | string a | Content/Json |
| POST | /Admin/RoxyFileman/ProcessRequest | ProcessRequest | string a | Content/Json |
| GET | /Admin/RoxyFileman/GetConfiguration | GetConfiguration | - | Json |

---

## 12. 店铺服务 (store)

### Admin - StoreController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Store | Index | - | Redirect to List |
| GET | /Admin/Store/List | List | - | View |
| POST | /Admin/Store/List | StoreList | StoreSearchModel searchModel | Json |
| GET | /Admin/Store/Create | Create | - | View |
| POST | /Admin/Store/Create | Create | StoreModel model, bool continueEditing | Redirect |
| GET | /Admin/Store/Edit/{id} | Edit | int id | View |
| POST | /Admin/Store/Edit | Edit | StoreModel model, bool continueEditing | Redirect |
| POST | /Admin/Store/Delete | Delete | int id | Redirect |

---

## 13. 安全服务 (security)

### Admin - SecurityController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Security/Permissions | Permissions | - | View |
| POST | /Admin/Security/PermissionsUpdate | PermissionsUpdate | PermissionMappingModel model | View |

### Admin - AuthenticationController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Authentication | Index | - | Redirect |
| GET | /Admin/Authentication/List | List | - | View |
| POST | /Admin/Authentication/List | AuthenticationProviderList | AuthenticationProviderSearchModel searchModel | Json |
| POST | /Admin/Authentication/ProviderUpdate | ProviderUpdate | AuthenticationProviderModel model | NullJson |

---

## 14. 本地化服务 (localization)

### Admin - LanguageController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Language | Index | - | Redirect to List |
| GET | /Admin/Language/List | List | - | View |
| POST | /Admin/Language/List | LanguageList | LanguageSearchModel searchModel | Json |
| GET | /Admin/Language/Create | Create | - | View |
| POST | /Admin/Language/Create | Create | LanguageModel model, bool continueEditing | Redirect |
| GET | /Admin/Language/Edit/{id} | Edit | int id | View |
| POST | /Admin/Language/Edit | Edit | LanguageModel model, bool continueEditing | Redirect |
| POST | /Admin/Language/Delete | Delete | int id | Redirect |
| POST | /Admin/Language/Resources | Resources | LocaleResourceSearchModel searchModel | Json |
| POST | /Admin/Language/ResourceUpdate | ResourceUpdate | LocaleResourceModel model | NullJson |
| POST | /Admin/Language/ResourceAdd | ResourceAdd | LocaleResourceModel model | Json |
| POST | /Admin/Language/ResourceDelete | ResourceDelete | int id | NullJson |
| POST | /Admin/Language/ExportCsv | ExportCsv | int id | File(CSV) |
| POST | /Admin/Language/ImportCsv | ImportCsv | int id, IFormFile importcsvfile | Redirect |

---

## 15. SEO服务 (seo)

### Admin - UrlRecordController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/UrlRecord | Index | - | Redirect to List |
| GET | /Admin/UrlRecord/List | List | - | View |
| POST | /Admin/UrlRecord/List | UrlRecordList | UrlRecordSearchModel searchModel | Json |
| POST | /Admin/UrlRecord/Update | Update | UrlRecordModel model | NullJson |
| POST | /Admin/UrlRecord/Delete | Delete | int id | NullJson |

---

## 16. GDPR服务 (gdpr)

### Admin - GdprController (隐含在CustomerController中)
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Gdpr | Index | - | Redirect |
| GET | /Admin/Gdpr/List | List | - | View |
| POST | /Admin/Gdpr/List | GdprLogList | GdprLogSearchModel searchModel | Json |
| POST | /Admin/Gdpr/Delete | GdprDelete | int id | Redirect |
| GET | /Admin/Gdpr/Export | GdprExport | int id | File(Excel) |

---

## 17. 联盟服务 (affiliate)

### Admin - AffiliateController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Affiliate | Index | - | Redirect to List |
| GET | /Admin/Affiliate/List | List | - | View |
| POST | /Admin/Affiliate/List | AffiliateList | AffiliateSearchModel searchModel | Json |
| GET | /Admin/Affiliate/Create | Create | - | View |
| POST | /Admin/Affiliate/Create | Create | AffiliateModel model, bool continueEditing | Redirect |
| GET | /Admin/Affiliate/Edit/{id} | Edit | int id | View |
| POST | /Admin/Affiliate/Edit | Edit | AffiliateModel model, bool continueEditing | Redirect |
| POST | /Admin/Affiliate/Delete | Delete | int id | Redirect |
| GET | /Admin/Affiliate/AddCustomerToAffiliatePopup | AddCustomerToAffiliatePopup | int affiliateId | View |
| POST | /Admin/Affiliate/AddCustomerToAffiliatePopup | AddCustomerToAffiliatePopup | - | Json(list) |

---

## 18. 目录服务 (directory)

### 前端 - CountryController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /country/getstatesbycountryid | GetStatesByCountryId | int countryId, bool addSelectStateItem | Json |

### Admin - CountryController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Country | Index | - | Redirect to List |
| GET | /Admin/Country/List | List | - | View |
| POST | /Admin/Country/List | CountryList | CountrySearchModel searchModel | Json |
| GET | /Admin/Country/Create | Create | - | View |
| POST | /Admin/Country/Create | Create | CountryModel model, bool continueEditing | Redirect |
| GET | /Admin/Country/Edit/{id} | Edit | int id | View |
| POST | /Admin/Country/Edit | Edit | CountryModel model, bool continueEditing | Redirect |
| POST | /Admin/Country/Delete | Delete | int id | Redirect |
| POST | /Admin/Country/StateList | StateList | StateProvinceSearchModel searchModel | Json |
| POST | /Admin/Country/StateUpdate | StateUpdate | StateProvinceModel model | NullJson |
| POST | /Admin/Country/StateDelete | StateDelete | int id | NullJson |

### Admin - CurrencyController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Currency | Index | - | Redirect to List |
| GET | /Admin/Currency/List | List | - | View |
| POST | /Admin/Currency/List | CurrencyList | CurrencySearchModel searchModel | Json |
| GET | /Admin/Currency/Create | Create | - | View |
| POST | /Admin/Currency/Create | Create | CurrencyModel model, bool continueEditing | Redirect |
| GET | /Admin/Currency/Edit/{id} | Edit | int id | View |
| POST | /Admin/Currency/Edit | Edit | CurrencyModel model, bool continueEditing | Redirect |
| POST | /Admin/Currency/Delete | Delete | int id | Redirect |
| GET | /Admin/Currency/LiveRates | LiveRates | - | View |
| POST | /Admin/Currency/ApplyRate | ApplyRate | int currencyId | Redirect |
| POST | /Admin/Currency/SaveRates | SaveRates | - | Redirect |

### Admin - MeasureController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Measure | Index | - | Redirect |
| GET | /Admin/Measure/Dimensions | Dimensions | - | View |
| POST | /Admin/Measure/Dimensions | MeasureDimensionList | MeasureDimensionSearchModel searchModel | Json |
| POST | /Admin/Measure/MeasureDimensionUpdate | MeasureDimensionUpdate | MeasureDimensionModel model | NullJson |
| POST | /Admin/Measure/MeasureDimensionAdd | MeasureDimensionAdd | MeasureDimensionModel model | Json |
| POST | /Admin/Measure/MeasureDimensionDelete | MeasureDimensionDelete | int id | NullJson |
| GET | /Admin/Measure/Weights | Weights | - | View |
| POST | /Admin/Measure/Weights | MeasureWeightList | MeasureWeightSearchModel searchModel | Json |
| POST | /Admin/Measure/MeasureWeightUpdate | MeasureWeightUpdate | MeasureWeightModel model | NullJson |
| POST | /Admin/Measure/MeasureWeightAdd | MeasureWeightAdd | MeasureWeightModel model | Json |
| POST | /Admin/Measure/MeasureWeightDelete | MeasureWeightDelete | int id | NullJson |

---

## 19. 日志服务 (logging)

### Admin - ActivityLogController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/ActivityLog | Index | - | Redirect |
| GET | /Admin/ActivityLog/ListTypes | ListTypes | - | View |
| POST | /Admin/ActivityLog/SaveTypes | SaveTypes | ActivityLogTypeModel model | View |
| GET | /Admin/ActivityLog/ListLogs | ListLogs | - | View |
| POST | /Admin/ActivityLog/ListLogs | ActivityLogList | ActivityLogSearchModel searchModel | Json |
| POST | /Admin/ActivityLog/Delete | Delete | int id | NullJson |
| POST | /Admin/ActivityLog/DeleteAll | DeleteAll | - | NullJson |

### Admin - LogController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Log | Index | - | Redirect to List |
| GET | /Admin/Log/List | List | - | View |
| POST | /Admin/Log/List | LogList | LogSearchModel searchModel | Json |
| POST | /Admin/Log/Delete | Delete | int id | NullJson |
| POST | /Admin/Log/DeleteAll | DeleteAll | - | NullJson |

---

## 20. 插件服务 (plugin)

### Admin - PluginController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Plugin/List | List | - | View |
| POST | /Admin/Plugin/Install | Install | string systemName | Redirect |
| POST | /Admin/Plugin/Uninstall | Uninstall | string systemName | Redirect |
| POST | /Admin/Plugin/Reload | Reload | string systemName | Redirect |
| POST | /Admin/Plugin/ReloadList | ReloadList | - | Redirect |
| GET | /Admin/Plugin/Edit | Edit | string systemName | View |
| POST | /Admin/Plugin/Edit | Edit | IFormCollection form | View |
| GET | /Admin/Plugin/OfficialFeed | OfficialFeed | - | View |
| POST | /Admin/Plugin/OfficialFeed | OfficialFeed | PluginSearchModel searchModel | Json |

---

## 21. 购物车服务 (cart)

### 前端 - ShoppingCartController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /cart | Cart | - | View |
| POST | /cart | Cart | ShoppingCartModel model | View/Redirect |
| POST | /cart/updatecart | UpdateCart | IFormCollection form | Redirect |
| POST | /cart/addproducttocart | AddProductToCart | int productId, int shoppingCartTypeId | Json/Redirect |
| POST | /cart/addproducttocart/details | AddProductToCartDetails | int productId, int shoppingCartTypeId | Json/Redirect |
| GET | /cart/estimateshippingpopup | EstimateShippingPopup | - | PartialView |
| POST | /cart/estimateshippingpopup | EstimateShippingPopup | EstimateShippingModel model | PartialView |
| POST | /cart/applydiscountcoupon | ApplyDiscountCoupon | string discountcouponcode | Redirect |
| POST | /cart/applygiftcard | ApplyGiftCard | string giftcardcouponcode | Redirect |
| POST | /cart/removeitemfromcart | RemoveItemFromCart | int id | Redirect |
| POST | /cart/removeitemfromcart/ajax | RemoveItemFromCartAjax | int id | Json |
| GET | /wishlist/{customerGuid?} | Wishlist | Guid? customerGuid | View |
| POST | /wishlist | Wishlist | WishlistModel model | View/Redirect |
| POST | /wishlist/additemtocart | AddItemToCart | int id | Redirect |
| POST | /wishlist/additemtocart/ajax | AddItemToCartAjax | int id | Json |
| POST | /wishlist/updatecart | UpdateWishlist | IFormCollection form | Redirect |
| POST | /wishlist/emailwishlist | EmailWishlist | WishlistEmailAFriendModel model | View |
| POST | /cart/uploadfile | UploadFile | int productId | Json |
| POST | /cart/uploadfileattribute | UploadFileAttribute | int attributeId | Json |

### Admin - ShoppingCartController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/ShoppingCart | Index | - | Redirect to List |
| GET | /Admin/ShoppingCart/List | List | - | View |
| POST | /Admin/ShoppingCart/List | ShoppingCartList | ShoppingCartSearchModel searchModel | Json |

---

## 22. 设置服务 (setting)

### Admin - SettingController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Setting/AllSettings | AllSettings | - | View |
| POST | /Admin/Setting/AllSettings | SettingList | SettingSearchModel searchModel | Json |
| POST | /Admin/Setting/SettingUpdate | SettingUpdate | SettingModel model | NullJson |
| POST | /Admin/Setting/SettingAdd | SettingAdd | SettingModel model | Json |
| POST | /Admin/Setting/SettingDelete | SettingDelete | int id | NullJson |
| GET | /Admin/Setting/Blog | Blog | - | View |
| POST | /Admin/Setting/Blog | Blog | BlogSettingsModel model | View |
| GET | /Admin/Setting/Vendor | Vendor | - | View |
| POST | /Admin/Setting/Vendor | Vendor | VendorSettingsModel model | View |
| GET | /Admin/Setting/Forum | Forum | - | View |
| POST | /Admin/Setting/Forum | Forum | ForumSettingsModel model | View |
| GET | /Admin/Setting/News | News | - | View |
| POST | /Admin/Setting/News | News | NewsSettingsModel model | View |
| GET | /Admin/Setting/Shipping | Shipping | - | View |
| POST | /Admin/Setting/Shipping | Shipping | ShippingSettingsModel model | View |
| GET | /Admin/Setting/Tax | Tax | - | View |
| POST | /Admin/Setting/Tax | Tax | TaxSettingsModel model | View |
| GET | /Admin/Setting/Catalog | Catalog | - | View |
| POST | /Admin/Setting/Catalog | Catalog | CatalogSettingsModel model | View |
| GET | /Admin/Setting/ShoppingCart | ShoppingCart | - | View |
| POST | /Admin/Setting/ShoppingCart | ShoppingCart | ShoppingCartSettingsModel model | View |
| GET | /Admin/Setting/Order | Order | - | View |
| POST | /Admin/Setting/Order | Order | OrderSettingsModel model | View |
| GET | /Admin/Setting/Gdpr | Gdpr | - | View |
| POST | /Admin/Setting/Gdpr | Gdpr | GdprSettingsModel model | View |
| GET | /Admin/Setting/Media | Media | - | View |
| POST | /Admin/Setting/Media | Media | MediaSettingsModel model | View |
| GET | /Admin/Setting/CustomerUser | CustomerUser | - | View |
| POST | /Admin/Setting/CustomerUser | CustomerUser | CustomerUserSettingsModel model | View |
| GET | /Admin/Setting/GeneralCommon | GeneralCommon | - | View |
| POST | /Admin/Setting/GeneralCommon | GeneralCommon | GeneralCommonSettingsModel model | View |
| GET | /Admin/Setting/ProductEditor | ProductEditor | - | View |
| POST | /Admin/Setting/ProductEditor | ProductEditor | ProductEditorSettingsModel model | View |
| GET | /Admin/Setting/Appearance | Appearance | - | View |
| POST | /Admin/Setting/Appearance | Appearance | AppearanceSettingsModel model | View |
| GET | /Admin/Setting/Minification | Minification | - | View |
| POST | /Admin/Setting/Minification | Minification | MinificationSettingsModel model | View |
| GET | /Admin/Setting/Ai | Ai | - | View |
| POST | /Admin/Setting/Ai | Ai | AiSettingsModel model | View |

---

## 23. 报表服务 (report)

### Admin - ReportController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Report/LowStock | LowStock | - | View |
| POST | /Admin/Report/LowStock | LowStockList | LowStockProductSearchModel searchModel | Json |
| GET | /Admin/Report/Bestsellers | Bestsellers | - | View |
| POST | /Admin/Report/Bestsellers | BestsellersList | BestsellerSearchModel searchModel | Json |
| GET | /Admin/Report/NeverSold | NeverSold | - | View |
| POST | /Admin/Report/NeverSold | NeverSoldList | NeverSoldSearchModel searchModel | Json |
| GET | /Admin/Report/CountrySales | CountrySales | - | View |
| POST | /Admin/Report/CountrySales | CountrySalesList | CountryReportSearchModel searchModel | Json |
| GET | /Admin/Report/RegisteredCustomers | RegisteredCustomers | - | View |
| POST | /Admin/Report/RegisteredCustomers | RegisteredCustomersList | RegisteredCustomersReportSearchModel searchModel | Json |

---

## 24. 公共服务/通用 (common/utility)

### 前端 - CommonController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /page-not-found | PageNotFound | - | View |
| GET | /language/setlanguage/{langid} | SetLanguage | int langid, string returnUrl | Redirect |
| GET | /currency/setcurrency/{currencyid} | SetCurrency | int currencyid, string returnUrl | Redirect |
| GET | /tax/settaxtype/{taxtype} | SetTaxType | int taxtype, string returnUrl | Redirect |
| GET | /contactus | ContactUs | - | View |
| POST | /contactus | ContactUs | ContactUsModel model | View/Redirect |
| GET | /contactvendor/{vendorId} | ContactVendor | int vendorId | View |
| POST | /contactvendor | ContactVendor | ContactVendorModel model | View/Redirect |
| GET | /sitemap | Sitemap | - | View |
| GET | /sitemap.xml | SitemapXml | int? id | File(XML) |
| GET | /storetheme/settheme/{themename} | SetStoreTheme | string themename | Redirect |
| POST | /eucookielawaccept | EuCookieLawAccept | - | Json |
| GET | /robots.txt | RobotsTextFile | - | Content |
| GET | /{genericSeName} | GenericUrl | - | View/Redirect |
| GET | /storeclosed | StoreClosed | - | View |
| GET | /internalredirect | InternalRedirect | string url, bool isPermanent | Redirect |
| GET | /fallbackredirect | FallbackRedirect | - | Redirect |

### 前端 - HomeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | / | Index | - | View |

### 前端 - ErrorController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /error | Error | - | File |

### 前端 - KeepAliveController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /keepalive | Index | - | Content |

### Admin - HomeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin | Index | - | View |

### Admin - CommonController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Common/Languages | Languages | - | View |
| POST | /Admin/Common/Languages | Languages | LanguageSearchModel searchModel | Json |
| GET | /Admin/Common/Currencies | Currencies | - | View |
| POST | /Admin/Common/Currencies | Currencies | CurrencySearchModel searchModel | Json |
| POST | /Admin/Common/SelectedCurrency | SelectedCurrency | int currencyId | NullJson |

### Admin - MenuController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Menu | Index | - | View |

### Admin - SearchCompleteController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/SearchComplete | SearchComplete | string term | Json |

---

## 25. 模板服务 (template)

### Admin - TemplateController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Template | Index | - | Redirect to List |
| GET | /Admin/Template/List | List | - | View |
| POST | /Admin/Template/List | TemplateList | TemplatesSearchModel searchModel | Json |

---

## 26. 地址属性服务 (address-attribute)

### Admin - AddressAttributeController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/AddressAttribute | Index | - | Redirect to List |
| GET | /Admin/AddressAttribute/List | List | - | View |
| POST | /Admin/AddressAttribute/List | AddressAttributeList | AddressAttributeSearchModel searchModel | Json |
| GET | /Admin/AddressAttribute/Create | Create | - | View |
| POST | /Admin/AddressAttribute/Create | Create | AddressAttributeModel model, bool continueEditing | Redirect |
| GET | /Admin/AddressAttribute/Edit/{id} | Edit | int id | View |
| POST | /Admin/AddressAttribute/Edit | Edit | AddressAttributeModel model, bool continueEditing | Redirect |
| POST | /Admin/AddressAttribute/Delete | Delete | int id | Redirect |
| POST | /Admin/AddressAttribute/ValueList | ValueList | AddressAttributeValueSearchModel searchModel | Json |
| POST | /Admin/AddressAttribute/ValueUpdate | ValueUpdate | AddressAttributeValueModel model | NullJson |
| POST | /Admin/AddressAttribute/ValueDelete | ValueDelete | int id | NullJson |
| GET | /Admin/AddressAttribute/ValueCreatePopup | ValueCreatePopup | int addressAttributeId | View |
| POST | /Admin/AddressAttribute/ValueCreatePopup | ValueCreatePopup | AddressAttributeValueModel model | Json |
| GET | /Admin/AddressAttribute/ValueEditPopup | ValueEditPopup | int id | View |
| POST | /Admin/AddressAttribute/ValueEditPopup | ValueEditPopup | AddressAttributeValueModel model | Json |

---

## 27. 小部件服务 (widget)

### Admin - WidgetController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/Widget | Index | - | Redirect |
| GET | /Admin/Widget/List | List | - | View |
| POST | /Admin/Widget/List | WidgetList | WidgetSearchModel searchModel | Json |
| POST | /Admin/Widget/WidgetUpdate | WidgetUpdate | WidgetModel model | NullJson |

---

## 28. 调度服务 (scheduler)

### 前端 - ScheduleTaskController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| POST | /scheduletask/runtask | RunTask | string taskType | - |

### Admin - ScheduleTaskController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/ScheduleTask | Index | - | Redirect to List |
| GET | /Admin/ScheduleTask/List | List | - | View |
| POST | /Admin/ScheduleTask/List | ScheduleTaskList | ScheduleTaskSearchModel searchModel | Json |
| POST | /Admin/ScheduleTask/TaskUpdate | TaskUpdate | ScheduleTaskModel model | NullJson |

---

## 29. 安装服务 (install)

### 前端 - InstallController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /install | Index | - | View |
| POST | /install | Index | InstallModel model | View/Redirect |
| GET | /install/changelanguage | ChangeLanguage | int languageId | Redirect |
| POST | /install/restartinstall | RestartInstall | - | Redirect |
| GET | /install/restartapplication | RestartApplication | - | Redirect |

---

## 30. AI服务 (ai)

### Admin - ArtificialIntelligenceController
| HTTP | Route | Method | Parameters | Response |
|------|-------|--------|------------|----------|
| GET | /Admin/ArtificialIntelligence | Index | - | Redirect |
| GET | /Admin/ArtificialIntelligence/GenerateText | GenerateText | - | View |
| POST | /Admin/ArtificialIntelligence/GenerateText | GenerateText | GenerateTextModel model | View |

---

## 统计汇总

| 微服务模块 | 前端路由数 | 管理端路由数 | 合计 |
|-----------|-----------|------------|------|
| 商品服务 (catalog) | 35 | 85+ | 120+ |
| 订单服务 (order) | 16 | 55+ | 71+ |
| 客户服务 (customer) | 25 | 45+ | 70+ |
| 支付服务 (payment) | 8 | 6 | 14 |
| 配送服务 (shipping) | 7 | 30 | 37 |
| 折扣服务 (discount) | 0 | 25 | 25 |
| 税务服务 (tax) | 0 | 7 | 7 |
| 供应商服务 (vendorsvc) | 5 | 16 | 21 |
| 内容服务 (content) | 30 | 45+ | 75+ |
| 消息服务 (message) | 9 | 30+ | 39+ |
| 媒体服务 (media) | 5 | 8 | 13 |
| 店铺服务 (store) | 0 | 8 | 8 |
| 安全服务 (security) | 0 | 4 | 4 |
| 本地化服务 (localization) | 0 | 12 | 12 |
| SEO服务 (seo) | 0 | 5 | 5 |
| GDPR服务 (gdpr) | 0 | 5 | 5 |
| 联盟服务 (affiliate) | 0 | 10 | 10 |
| 目录服务 (directory) | 1 | 20+ | 21+ |
| 日志服务 (logging) | 0 | 12 | 12 |
| 插件服务 (plugin) | 0 | 8 | 8 |
| 购物车服务 (cart) | 18 | 3 | 21 |
| 设置服务 (setting) | 0 | 30+ | 30+ |
| 报表服务 (report) | 0 | 10 | 10 |
| 公共服务 (common) | 16 | 5 | 21 |
| 模板服务 (template) | 0 | 3 | 3 |
| 地址属性服务 | 0 | 14 | 14 |
| 小部件服务 (widget) | 0 | 4 | 4 |
| 调度服务 (scheduler) | 1 | 4 | 5 |
| 安装服务 (install) | 5 | 0 | 5 |
| AI服务 (ai) | 0 | 2 | 2 |

**总计: 约 700+ 个API路由端点**

---

### 关键架构模式说明

1. **Admin区域路由统一前缀**: 所有管理端路由以 `/Admin/{Controller}/` 为前缀
2. **CRUD模式**: 大多数Admin控制器遵循 `List(GET) -> List(POST Json) -> Create(GET+POST) -> Edit(GET+POST) -> Delete(POST)` 标准模式
3. **内联编辑模式**: Admin列表页使用 `XxxUpdate(POST NullJson)` 和 `XxxDelete(POST NullJson)` 实现AJAX内联编辑/删除
4. **弹窗选择模式**: `XxxAddPopup(GET+POST)` 用于弹窗选择关联实体(如产品、分类、制造商)
5. **OPC (One-Page Checkout)**: 前端结账使用 `OpcSaveXxx(POST Json)` 系列方法实现单页结账AJAX交互
6. **权限控制**: Admin端方法使用 `[CheckPermission(StandardPermission.XXX.YYY)]` 属性进行细粒度权限控制
7. **文件导出**: 支持 XML/Excel/CSV/PDF 多种格式导出，使用 `ExportXxxAll(POST file)` 和 `ExportXxxSelected(POST file)` 模式