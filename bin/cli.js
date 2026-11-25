#!/usr/bin/env node

// The last npm release only points to the Go binary, which updates itself with `ccc update`.
const INSTALL_COMMAND = 'curl -fsSL https://raw.githubusercontent.com/Jerome1337/ccc/main/install.sh | sh';

console.error(`
@jerome1337/ccc is now a standalone binary that updates itself. Install it with:

  ${INSTALL_COMMAND}

Then run: ccc update --global
`);

process.exit(1);
