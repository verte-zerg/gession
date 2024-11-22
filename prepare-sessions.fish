#!/usr/bin/env fish

# Define sessions and their windows with commands
set sessions "system:monitor:top#disk-usage:df -h#processes:ps aux" \
    "blog:helix:hx ../blog#gitui:cd ../blog && gitui#deploy:cd ../blog && hugo server" \
    "qrcode:helix:cd ../qrcode && hx README.md#git:cd ../qrcode && gitui" \
    "gession:helix:hx README.md#git:gitui#log:tail -f '../../Library/Application Support/gession/gession.log'"

for session_entry in $sessions
    set session (echo $session_entry | cut -d: -f1)
    set windows (echo $session_entry | cut -d: -f2- | string split "#")

    set first_window_name (echo $windows | head -1 | cut -d: -f1)

    for window_entry in $windows
        set name (echo $window_entry | cut -d: -f1)
        set cmd (echo $window_entry | cut -d: -f2-)

        if not tmux has-session -t $session >/dev/null 2>&1
            tmux new-session -d -s $session -n $first_window_name
        else
            tmux new-window -t $session -n $name "$SHELL"
        end

        tmux send-keys -t $session:$name "$cmd" C-m
    end
end

echo "All sessions and windows have been configured."
