for f in $(ls -1 images); do
    p="images/$f"
    h=$(shasum -b $p | awk '{print $1}')
    mv $p images/$h.jpg
done
