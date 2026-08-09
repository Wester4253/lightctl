"""CLI entry points. Deliberately thin: parse args, call commands.py,
print the result. No business logic here.
"""

import typer

import commands
from config import load_config
from models import GoveeError, Profile

app = typer.Typer(help="Control your Govee lights via Home Assistant.")


def _run(func, *args):
    try:
        config = load_config()
        return func(config, *args)
    except GoveeError as exc:
        typer.echo(f"Error: {exc}", err=True)
        raise typer.Exit(code=1)


def _prompt_pc_action(profile: Profile) -> None:
    if not profile.pc_action_prompt:
        return
    typer.echo("\nGoodnight!")
    typer.echo("==============")
    typer.echo("1) Shut down")
    typer.echo("2) Sleep")
    typer.echo("3) Ignore")
    choice = typer.prompt("Choose")
    try:
        typer.echo(commands.run_pc_action(choice))
    except GoveeError as exc:
        typer.echo(f"Error: {exc}", err=True)


def _make_profile_command(name: str):
    def _command():
        profile = _run(commands.apply_profile, name)
        typer.echo(f"Activated '{name}'.")
        _prompt_pc_action(profile)

    _command.__doc__ = f"Run the '{name}' profile."
    return _command


# One top-level command per default profile: lightctl night / gaming / movie / work / relax
for _profile_name in ("night", "gaming", "movie", "work", "relax"):
    app.command(name=_profile_name)(_make_profile_command(_profile_name))


@app.command()
def profile(name: str):
    """Run any configured profile by name (including custom ones you add)."""
    profile_obj = _run(commands.apply_profile, name)
    typer.echo(f"Activated profile '{name}'.")
    _prompt_pc_action(profile_obj)


@app.command()
def effect(name: str):
    """Set the light effect (e.g. `lightctl effect Rainbow`)."""
    _run(commands.set_effect, name)
    typer.echo(f"Effect set to '{name}'.")


@app.command()
def brightness(pct: int):
    """Set brightness 0-100 (e.g. `lightctl brightness 50`)."""
    _run(commands.set_brightness, pct)
    typer.echo(f"Brightness set to {pct}%.")


@app.command()
def color(r: int, g: int, b: int):
    """Set RGB color (e.g. `lightctl color 255 0 255`)."""
    _run(commands.set_color, r, g, b)
    typer.echo(f"Color set to ({r}, {g}, {b}).")


@app.command()
def on():
    """Turn the light on."""
    _run(commands.turn_on)
    typer.echo("Light on.")


@app.command()
def off():
    """Turn the light off."""
    _run(commands.turn_off)
    typer.echo("Light off.")


@app.command()
def effects():
    """List effects currently available on the light."""
    state = _run(commands.get_light_state)
    if not state.available_effects:
        typer.echo("No effects reported.")
        return
    for name in state.available_effects:
        typer.echo(f"  {name}")


@app.command()
def status():
    """Show current light state."""
    state = _run(commands.get_light_state)
    typer.echo(f"Status: {'on' if state.is_on else 'off'}")
    typer.echo(f"Effect: {state.effect or '-'}")
    typer.echo(
        f"Brightness: {state.brightness_pct}%"
        if state.brightness_pct is not None
        else "Brightness: -"
    )


if __name__ == "__main__":
    app()
