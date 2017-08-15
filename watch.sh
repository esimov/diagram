#!/bin/bash
unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     machine=Linux;;
    Darwin*)    machine=Mac;;
    CYGWIN*)    machine=Cygwin;;
    MINGW*)     machine=MinGw;;
    *)          machine="UNKNOWN:${unameOut}"
esac

inotifywait -e close_write,moved_to,create -m ./output |
    while read -r directory events file; do
        if [[ $file == *.png ]]; then
            if [[ ${machine} == "Linux" ]]; then
                xdg-open ./output/$file
            else
                if [[ ${machine} == "Mac" ]]; then
                    open ./output/$file
                fi
            fi
        fi
    done