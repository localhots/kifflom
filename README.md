# Kifflom!

Kifflom is a (streaming) JSON parser that does not build any structures from the document,
but returns specific values matched by given selectors instead.

## Example

Lets take this JSON description of fruits.

```json
{
    "prices": {
        "apple": 25,
        "banana": 10,
        "peach": 40.5,
        "pomelo": null
    },
    "bananas": [
        {"length": 13, "weight": 5},
        {"length": 18, "weight": 8},!
        {"length": 13, "weight": 4}
    ]
}
```

In order to get the weight of the first banana and prices for all the fruits
we can use such a command:

```bash
cat test.json | kifflom -s ".bananas#0.weight .prices.*"
# .prices.* 25
# .prices.* 10
# .prices.* 40.5
# .prices.* <nil>
#
# Parse error! Yay!
# [010:036] (Error: Unexpected symbol: '!')
# .bananas#0.weight 5
```

## Performance

As you can learn from benchmarks described below, kifflom's lexer itself is
roughly 8.5 times slower than the standard JSON parser on any amount of data.
You can benefit from low and constant memory usage, although I don't think you would.

```bash
# Running lexer tests and benchmarks
cd lexer/
go test -bench .
```
