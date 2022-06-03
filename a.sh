progress(){
    local b=""
    for((i=0;i<=100;i++)); do
        printf "progress: [%-100s] %d%%\r" $b $i
        sleep 2.2
        b+="#"
    done
}

progress
