#!/usr/bin/env zsh
# zmosh-picker — single-keypress session launcher for zmosh
# https://github.com/nerveband/zmosh-picker

_zmosh_picker_main() {
  # ─── Guards ─────────────────────────────────────────────────────────
  [[ ! -o interactive ]] && return 0
  [[ -n "$ZMX_SESSION" && -z "$ZPICK" ]] && return 0
  [[ ! -t 0 ]] && return 0
  command -v zmosh &>/dev/null || return 0
  unset ZPICK

  # ─── Colors ─────────────────────────────────────────────────────────

  local R=$'\e[0m' D=$'\e[2m'
  local rd=$'\e[31m' cy=$'\e[36m' gn=$'\e[32m' yl=$'\e[33m' mg=$'\e[35m'
  local Brd=$'\e[1;31m' Bcy=$'\e[1;36m' Bgn=$'\e[1;32m' Byl=$'\e[1;33m' Bwh=$'\e[1;97m'

  # ─── Name generators ───────────────────────────────────────────────

  # Counter: bare <dirname> when available, then <dirname>-2, -3, …
  _zmosh_pick_name_counter() {
    local dirname="${${1:-$PWD}:t}"
    if (( ! ${session_names[(Ie)${dirname}]} )); then
      print -r -- "${dirname}"
    else
      local n=2
      while (( ${session_names[(Ie)${dirname}-${n}]} )); do
        (( n++ ))
      done
      print -r -- "${dirname}-${n}"
    fi
  }

  # Date: <dirname>-MMDD (zsh builtin, no date subprocess)
  _zmosh_pick_name_date() {
    local dirname="${${1:-$PWD}:t}"
    zmodload -F zsh/datetime b:strftime
    local ds
    strftime -s ds '%m%d' "$EPOCHSECONDS"
    print -r -- "${dirname}-${ds}"
  }

  # ─── Main loop (re-parses sessions each iteration) ─────────────────

  while true; do

  # ─── Parse active sessions (single zmosh call) ─────────────────────

  local -a session_names session_clients session_dirs
  local line name clients dir field is_home
  local -a parts

  session_names=() session_clients=() session_dirs=()

  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    name="" clients="" dir=""
    for field in ${(s:	:)line}; do
      case "$field" in
        session_name=*) name="${field#session_name=}" ;;
        clients=*) clients="${field#clients=}" ;;
        started_in=*) dir="${field#started_in=}" ;;
      esac
    done
    [[ -n "$name" ]] || continue
    session_names+=("$name")
    session_clients+=("$clients")
    is_home=0
    [[ "$dir" == "$HOME"* ]] && is_home=1
    dir="${dir/$HOME/~}"
    parts=(${(s:/:)dir})
    if (( ${#parts} > 4 )); then
      if (( is_home )); then
        dir="~/${parts[-3]}/${parts[-2]}/${parts[-1]}"
      else
        dir=".../${parts[-3]}/${parts[-2]}/${parts[-1]}"
      fi
    fi
    session_dirs+=("$dir")
  done < <(zmosh list 2>/dev/null)

  local session_count=${#session_names}

  # ─── Build key-to-session mapping ──────────────────────────────────

  local -A key_to_session
  local -a keys_display
  local key_chars="123456789abdefghijlmnopqrstuvwxy"
  local i key

  key_to_session=()
  keys_display=()

  for (( i=1; i<=session_count && i<=${#key_chars}; i++ )); do
    key="${key_chars[$i]}"
    key_to_session[$key]="${session_names[$i]}"
    keys_display+=("$key")
  done

  # ─── Display ────────────────────────────────────────────────────────

  local default_name
  default_name="$(_zmosh_pick_name_counter)"

  echo ""

  if (( session_count > 0 )); then
    printf "  ${Bcy}zmosh${R} ${D}${session_count} session%s${R}\n" \
      "$( (( session_count > 1 )) && echo s)"
    echo ""

    local cc
    for (( i=1; i<=session_count; i++ )); do
      if (( session_clients[$i] > 0 )); then
        cc="${Bgn}*${R}"
      else
        cc="${D}.${R}"
      fi
      printf "  ${Byl}%s${R}  ${Bwh}%s${R} ${cc} ${D}%s${R}\n" \
        "${keys_display[$i]}" \
        "${session_names[$i]}" \
        "${session_dirs[$i]}"
    done

    echo ""
  else
    printf "  ${Bcy}zmosh${R} ${D}no sessions${R}\n\n"
  fi

  printf "  ${Bgn}enter${R} ${D}new${R} ${Bwh}%s${R}\n" "$default_name"
  printf "  ${mg}c${R} ${D}custom${R}  ${mg}z${R} ${D}pick dir${R}  ${cy}d${R} ${D}+date${R}  ${rd}k${R} ${D}kill${R}  ${yl}esc${R} ${D}skip${R}\n"
  echo ""

  # ─── Single-keypress input ─────────────────────────────────────────

  printf "  ${Bcy}>${R} "
  local choice
  read -k1 choice 2>/dev/null
  echo ""

  case "$choice" in
    $'\e')
      return 0
      ;;
    $'\n')
      printf "\n  ${Bgn}>${R} ${Bwh}%s${R}\n\n" "$default_name"
      exec zmosh attach "$default_name"
      ;;
    d)
      local date_name
      date_name="$(_zmosh_pick_name_date)"
      printf "\n  ${Bgn}>${R} ${Bwh}%s${R}\n\n" "$date_name"
      exec zmosh attach "$date_name"
      ;;
    c)
      printf "\n  ${mg}name:${R} "
      local custom_name
      read -r custom_name 2>/dev/null
      if [[ -z "$custom_name" ]]; then
        continue
      fi
      printf "\n  ${Bgn}enter${R} ${D}create in ~${R}  ${mg}z${R} ${D}pick dir${R}  ${yl}esc${R} ${D}cancel${R}\n\n"
      printf "  ${Bcy}>${R} "
      local csub
      read -k1 csub 2>/dev/null
      echo ""
      case "$csub" in
        $'\n')
          printf "\n  ${Bgn}>${R} ${Bwh}%s${R}\n\n" "$custom_name"
          exec zmosh attach "$custom_name"
          ;;
        z)
          if command -v zoxide &>/dev/null; then
            echo ""
            local cpicked_dir
            cpicked_dir="$(zoxide query -i 2>/dev/null)"
            if [[ -n "$cpicked_dir" ]]; then
              cd "$cpicked_dir" || true
              printf "\n  ${Bgn}>${R} ${Bwh}%s${R} ${D}%s${R}\n\n" "$custom_name" "$cpicked_dir"
              exec zmosh attach "$custom_name"
            else
              continue
            fi
          else
            printf "  ${yl}zoxide not installed${R}\n"
            continue
          fi
          ;;
        *)
          continue
          ;;
      esac
      ;;
    k)
      if (( session_count == 0 )); then
        printf "  ${D}no sessions to kill${R}\n"
        continue
      fi
      printf "\n  ${Brd}kill${R} ${D}which session?${R} "
      local kill_choice
      read -k1 kill_choice 2>/dev/null
      echo ""
      if [[ -n "${key_to_session[$kill_choice]}" ]]; then
        local kill_target="${key_to_session[$kill_choice]}"
        local confirm_kill="${ZMOSH_PICKER_NO_CONFIRM:-0}"
        if [[ "$confirm_kill" != "1" ]]; then
          printf "  ${Brd}kill ${Bwh}%s${Brd}?${R} ${D}y/n${R} " "$kill_target"
          local yn
          read -k1 yn 2>/dev/null
          echo ""
          if [[ "$yn" != "y" && "$yn" != "Y" ]]; then
            printf "  ${D}cancelled${R}\n"
            continue
          fi
        fi
        zmosh kill "$kill_target" 2>/dev/null
        printf "  ${Brd}killed${R} ${Bwh}%s${R}\n" "$kill_target"
      else
        printf "  ${D}cancelled${R}\n"
      fi
      continue
      ;;
    z)
      if command -v zoxide &>/dev/null; then
        echo ""
        local picked_dir
        picked_dir="$(zoxide query -i 2>/dev/null)"
        if [[ -n "$picked_dir" ]]; then
          cd "$picked_dir" || true
          local zname
          zname="$(_zmosh_pick_name_counter "$picked_dir")"
          printf "\n  ${Bgn}>${R} ${Bwh}%s${R} ${D}%s${R}\n\n" "$zname" "$picked_dir"
          exec zmosh attach "$zname"
        else
          continue
        fi
      else
        printf "  ${yl}zoxide not installed${R}\n"
        continue
      fi
      ;;
    *)
      if [[ -n "${key_to_session[$choice]}" ]]; then
        local target="${key_to_session[$choice]}"
        printf "\n  ${Bgn}>${R} ${Bwh}%s${R}\n\n" "$target"
        exec zmosh attach "$target"
      else
        return 0
      fi
      ;;
  esac

  done
}

_zmosh_picker_main
unfunction _zmosh_picker_main _zmosh_pick_name_counter _zmosh_pick_name_date 2>/dev/null
