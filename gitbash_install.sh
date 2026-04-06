pjvm() {
    do_eval="false"
    #echo "$@"
    while IFS= read -r line; do
        #echo $do_eval $line
        if [ "$line" = "@@@START_SHELL@@@" ]; then
            do_eval="true"
        elif [ "$line" = "@@@END_SHELL@@@" ]; then
            do_eval="false"
        elif [ "$do_eval" = "true" ]; then
           #echo "would eval $line"
           eval "$line"
        else
            echo "$line"
        fi
    done < <("@@@PJVM_EXEC@@@" -shell GitBash "$@")
}