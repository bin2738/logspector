#!/usr/bin/env bash

# Скрипт автодополнения для утилиты logspector

_logspector_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Все доступные флаги утилиты
    opts="-c -f -l -json -q -v -no-color -since -until -h"

    # Если предыдущий флаг был -l, предлагаем уровни логов
    if [[ ${prev} == "-l" ]]; then
        COMPREPLY=( $(compgen -W "ERROR WARN INFO DEBUG" -- ${cur}) )
        return 0
    fi

    # Если предыдущий флаг требует путь к файлу (-c или -f), предлагаем файлы из текущей директории
    if [[ ${prev} == "-c" || ${prev} == "-f" ]]; then
        COMPREPLY=( $(compgen -f -- ${cur}) )
        return 0
    fi

    # Автодополнение самих флагов (если пользователь вводит "-")
    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}

complete -F _logspector_completion logspector