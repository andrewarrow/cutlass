-- Define the hotkey (Cmd + Option + M)
hs.hotkey.bind({"cmd", "alt"}, "M", function()
    -- Get all open windows
    local windows = hs.window.allWindows()
    
    -- Loop through each window and minimize it
    for _, window in ipairs(windows) do
        if window:application():name() ~= "Hammerspoon" then -- Skip Hammerspoon's own windows
            window:minimize()
        end
    end
    
    -- Optional: Notify user the action was completed
    hs.notify.new({title="Hammerspoon", informativeText="All windows minimized"}):send()
end)


hs.hotkey.bind({"cmd", "alt", "ctrl"}, "R", function()
    -- Launch QuickTime if it's not running
    hs.application.launchOrFocus("QuickTime Player")
    hs.timer.doAfter(2, function()
        -- Open New Screen Recording
        hs.eventtap.keyStroke({"cmd", "ctrl"}, "n", 0)
        
        hs.timer.doAfter(2, function()
        hs.eventtap.keyStroke({}, "return", 0)
                    -- Open New Movie Recording
                    hs.eventtap.keyStroke({"cmd", "alt"}, "n", 0)
                    
                    hs.timer.doAfter(2, function()
                        -- Start movie recording by clicking the red record button
                        hs.eventtap.keyStroke({}, "space", 0)
            end)
        end)
    end)
end)

