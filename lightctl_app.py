"""Internal entry point used by the lightctl launcher."""

import sys


def main() -> None:
    if len(sys.argv) == 1:
        from tui import GoveeApp

        GoveeApp().run()
    else:
        from cli import app

        sys.argv[0] = "lightctl"
        app()


if __name__ == "__main__":
    main()
