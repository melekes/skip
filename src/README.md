# Skip challenge

## Assumptions

1. If a token has zero attributes, it's rarity is `0.0`
2. Maximum 3 attempts are made to fetch a single token (configurable; see `MAX_RETRY`)
3. Maximum 250 goroutines are run to fetch all tokens (configurable; see `MAX_WORKERS`)

## More things to consider / do

1. How large the attributes set could be?
2. Exponential backoff for retries (especially if the server has rate limiting)
3. Make rarity calculation concurrent using `sync.Map` instead of `Map` (if we want more speed)

## External libraries used

### Testing

- https://github.com/h2non/gock for mocking Go's http
- https://github.com/stretchr/testify for asserts
