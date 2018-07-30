for f in $(ls -1 ./images); do
    echo "GET http://localhost:8080/convert/e/imaging/w/50/h/50/$f"
    echo "GET http://localhost:8080/convert/e/imaging/w/250/h/250/$f"
    echo "GET http://localhost:8080/convert/e/imaging/w/500/h/500/$f"
    echo "GET http://localhost:8080/convert/e/imaging/w/640/h/480/$f"
    echo "GET http://localhost:8080/convert/e/imaging/w/1000/h/1000/$f"
done
