def binds_and_header(mapping, emoji="🐈"):
    bind_parts = []
    header_parts = []
    for keybind, (prompt, command) in mapping.items():
        header_parts.append(f"{keybind}: {prompt}")
        bind = f"{keybind}:change-prompt({emoji} {prompt} > )+reload({command})"
        if "'" in bind:
            raise ValueError("can't have ' in bind")
        bind_parts.append(bind)
    binds = ','.join(bind_parts)
    header = ' | '.join(header_parts)
    return binds, header

