curl "https://unsplash.com/napi/photos?per_page=30" | jq '.[].urls.full' -r | xargs -n1 wget
