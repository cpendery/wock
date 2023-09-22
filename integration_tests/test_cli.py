import requests
import pytest
import http
import subprocess
import platform
from os import path

html_dir = path.join("integration_tests", "html")


@pytest.fixture
def command():
    plat = platform.system()
    return "./wock" if plat == "Linux" or plat == "Darwin" else "wock.exe"


@pytest.fixture
def admin_command():
    plat = platform.system()
    return ["sudo", "./wock"] if plat == "Linux" or plat == "Darwin" else ["wock.exe"]


@pytest.mark.parametrize(
    ["args", "expected", "expected_status"],
    [
        pytest.param(["uninstall"], "Local CA was not installed", 0),
        pytest.param(["status"], "[offline]", 0),
        pytest.param(["stop"], "Daemon is already offline", 0),
        pytest.param(["rm", "nytimes.com"], "Daemon is offline, no hosts to remove", 0),
        pytest.param(
            ["nytimes.com", html_dir],
            "Error: local CA is not installed, run `wock install` to install the CA",
            1,
        ),
        pytest.param(
            ["clear"],
            "Error: local CA is not installed, run `wock install` to install the CA",
            1,
        ),
    ],
)
def test_commands_when_uninstalled(command, args, expected, expected_status):
    result = subprocess.run([command, *args], capture_output=True, text=True)
    assert result.returncode == expected_status
    if expected_status == 0:
        assert expected in result.stdout
    else:
        assert expected in result.stderr


def test_start(command, admin_command):
    result = subprocess.run([command, "status"], capture_output=True, text=True)
    assert "[offline]" in result.stdout

    subprocess.Popen(
        [*admin_command, "start"],
        shell=False,
        stdin=None,
        stdout=None,
        stderr=None,
        close_fds=True,
    )

    result = subprocess.run([command, "status"], capture_output=True, text=True)
    assert "[online]" in result.stdout


def test_install(admin_command):
    result = subprocess.run([*admin_command, "install"], capture_output=True, text=True)
    assert result.returncode == 0
    assert "Successfully installed/verified local CA" in result.stdout


def test_already_installed_install(command):
    result = subprocess.run([command, "install"], capture_output=True, text=True)
    assert result.returncode == 0
    assert "Local CA is already installed" in result.stdout


def test_host_successful(command, snapshot):
    result = subprocess.run(
        [command, "nytimes.com", html_dir], capture_output=True, text=True
    )
    assert result.returncode == 0
    assert "mocking host" in result.stdout

    res = requests.get("https://nytimes.com", verify=False)
    assert res.status_code == http.HTTPStatus.OK
    assert res.content == snapshot


@pytest.mark.parametrize(
    ["args", "expected"],
    [
        pytest.param(["nytimes.com", "host"], "unable to serve"),
        pytest.param(["nytimes.", html_dir], "is an invalid hostname"),
    ],
)
def test_host_failure(command, args, expected):
    result = subprocess.run([command, *args], capture_output=True, text=True)
    assert result.returncode == 1
    assert expected in result.stderr


def test_status_mocking_host(command):
    result = subprocess.run([command, "status"], capture_output=True, text=True)
    assert "nytimes.com" in result.stdout


def test_rm(command):
    assert (
        "nytimes.com"
        in subprocess.run([command, "status"], capture_output=True, text=True).stdout
    )
    result = subprocess.run(
        [command, "rm", "nytimes.com"], capture_output=True, text=True
    )
    assert result.returncode == 0

    assert (
        "nytimes.com"
        not in subprocess.run(
            [command, "status"], capture_output=True, text=True
        ).stdout
    )


def test_clear(command, snapshot):
    subprocess.run([command, "google.com", html_dir], capture_output=True, text=True)
    subprocess.run([command, "apple.com", html_dir], capture_output=True, text=True)

    mockedHost1 = requests.get("https://google.com", verify=False).content
    mockedHost2 = requests.get("https://apple.com", verify=False).content

    result = subprocess.run([command, "clear"], capture_output=True, text=True)
    assert "Successfully cleared all hosts" in result.stdout

    unmockedHost1 = requests.get("https://google.com", verify=False).content
    unmockedHost2 = requests.get("https://apple.com", verify=False).content

    assert mockedHost1 != unmockedHost1
    assert mockedHost2 != unmockedHost2
    assert mockedHost1 == snapshot
    assert mockedHost2 == snapshot


def test_stop(command):
    assert (
        "[online]"
        in subprocess.run([command, "status"], capture_output=True, text=True).stdout
    )
    assert (
        "Successfully stopped daemon"
        in subprocess.run([command, "stop"], capture_output=True, text=True).stdout
    )
    assert (
        "[offline]"
        in subprocess.run([command, "status"], capture_output=True, text=True).stdout
    )
