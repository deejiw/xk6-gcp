import interpret from 'k6/x/interpret';

export default function() {
    var result = interpret.run(
    `package interpret

    import "fmt"


    // fibonacci is a function that returns
    // a function that returns an int.
    func fibonacci() func() int {
        x, y := 0, 1
        return func() int {
            x, y = y, x + y
            fmt.Println("Calculated:",x)
            return x
        }
    }

    func Run(s interface{}) interface{} {
        f := fibonacci()
        var results []int
        var i int64
        for i = 0; i < s.(int64); i++ {
            results = append(results,f())
        }
        return results
    }
    `,
    10
    );

    console.log(`Results: ${result}`)
}
