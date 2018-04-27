package rx

type (
        // MapFunc defines a mapping function for the Map operator.
        MapFunc func(interface{}) interface{}

        // ScanFunc defines a scanning function for the Scan operator.
        ScanFunc func(interface{}, interface{}) interface{}

        // FilterFunc defines a filtering function for the Filter operator.
        FilterFunc func(interface{}) bool

	// IterFunc defines an iterating function for the ForEach operator.
	// IterFunc func(interface{})

	// FoldFunc is equivalent to ScanFunc, but used with Folding operators.
	// FoldFunc func(interface{}, interface{}) interface{}
)

type Functor interface {
	Map(MapFunc) Observable
	Filter(FilterFunc) Observable
	Scan(ScanFunc) Observable
	// ForEach(IterFunc)
	// Foldl(FoldFunc) interface{}
	// Foldr(FoldFunc) interface{}
}
