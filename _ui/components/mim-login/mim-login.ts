/* Login component */

declare var CryptoJS: any;

@Polymer.decorators.customElement('mim-login')
class MimLogin extends Polymer.Element {

  @Polymer.decorators.property({type: Boolean, notify: true})
  loggedIn: boolean;

  @Polymer.decorators.property({type: String})
  loginError: string;

  ready() {
    super.ready();
    this.$.username.addEventListener('keydown', this.keydown.bind(this));
    this.$.password.addEventListener('keydown', this.keydown.bind(this));
  }

  connectedCallback() {
    super.connectedCallback();
    setTimeout(() => {
      this.$.username.focus();
    }, 0);
  }

  async login() {
    const username = this.$.username.value;
    const password = this.$.password.value;
    const seconds = Math.floor(Date.now()/1000);
    console.log("in login: username=", username, ", password=", password, ", seconds=", seconds);
    const cryptword = this.sha256sum(username + "-" + password);
    console.log("cryptword:", cryptword);
    const shaInput = cryptword + "-" + seconds.toString();
    console.log("shaInput:", shaInput);
    const nonce = this.sha256sum(shaInput);
    console.log("nonce:", nonce);
    try {
      const loginUrl = "/auth/login/";
      const formData = new FormData();
      formData.append("userid", username);
      formData.append("nonce", nonce);
      formData.append("time", seconds.toString());
      const options = {
        method: "POST",
        params: formData,
      };
      const response = await ApiManager.xhrJson(loginUrl, options);
      this.loggedIn = true;
      console.log("Login succeeded");
      location.reload();
    } catch (e) {
      this.loggedIn = false;
      this.loginError = "Login failed";
    }
  }

  async logout() {
    try {
      const loginUrl = "/auth/logout/";
      const response = await ApiManager.xhrJson(loginUrl);
      this.loggedIn = false;
      console.log("Logout succeeded");
      location.reload();
    } catch (e) {
      console.error("Logout failed");
    }
  }

  // CheckStatus checks to see if we are logged in and sets our loggedIn flag
  // accordingly.
  async checkStatus() {
    try {
      const oldStatus = this.loggedIn;
      const statusUrl = "/auth/status/";
      const response = await ApiManager.xhrJson(statusUrl);
      this.loggedIn = response.LoggedIn;
      if (this.loggedIn != oldStatus && !this.loggedIn) {
        console.error("not logged in");
        location.reload();    // TODO - use a dialog to relogin without reload
      }
    } catch (e) {
      console.error("auth status call failed");
    }
  }

  keydown(e: any) {
    this.loginError = "";
    if (e.key == "Enter") {
      if (this.$.username.focused) {
        this.$.password.focus();
      } else if (this.$.password.focused) {
        this.login();
      }
    }
  }

  sha256sum(s: string) {
    const w = CryptoJS.SHA256(s);
    return w.toString();
  }
}
