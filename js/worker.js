export default {
  async fetch(request, env, ctx) {
    // Parse credentials from the request
    const url = new URL(request.url);
    const username = url.searchParams.get("username");
    const password = url.searchParams.get("password");

    if (!username || !password) {
      return new Response("Username and Password are required", {
        status: 400,
      });
    }

    const baseURL = "https://login.gitam.edu";
    const glearnURL = "https://glearn.gitam.edu";
    const loginPath = "/Login.aspx";
    const homePath = "/Student/std_course_details";

    try {
      // Step 1: Fetch the login page
      const loginPageResponse = await fetch(baseURL + loginPath, {
        method: "GET",
        headers: {
          "User-Agent":
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        },
      });
      const loginPageHTML = await loginPageResponse.text();

      // Step 2: Extract form values
      const viewState = extractHiddenValue(loginPageHTML, "__VIEWSTATE");
      const eventValidation = extractHiddenValue(
        loginPageHTML,
        "__EVENTVALIDATION",
      );
      const viewStateGenerator =
        extractHiddenValue(loginPageHTML, "__VIEWSTATEGENERATOR") || "C2EE9ABB";

      if (!viewState || !eventValidation) {
        throw new Error("Failed to extract necessary form values.");
      }

      // Step 3: Submit login form
      const formData = new URLSearchParams();
      formData.append("__EVENTTARGET", "");
      formData.append("__EVENTARGUMENT", "");
      formData.append("__VIEWSTATE", viewState);
      formData.append("__VIEWSTATEGENERATOR", viewStateGenerator);
      formData.append("__EVENTVALIDATION", eventValidation);
      formData.append("txtusername", username);
      formData.append("password", password);
      formData.append("Submit", "Login");

      const loginResponse = await fetch(baseURL + loginPath, {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          Referer: baseURL + loginPath,
          "User-Agent":
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        },
        body: formData,
        redirect: "manual",
      });

      if (loginResponse.status !== 302) {
        throw new Error("Login failed. Please check your credentials.");
      }

      // Step 4: Access the home page
      const homePageResponse = await fetch(glearnURL + homePath, {
        method: "GET",
        headers: {
          "User-Agent":
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
          Cookie: loginResponse.headers.get("set-cookie") || "",
        },
      });
      const homePageHTML = await homePageResponse.text();

      // Return the home page content
      return new Response(homePageHTML, {
        status: 200,
        headers: { "Content-Type": "text/html" },
      });
    } catch (err) {
      return new Response(`Error: ${err.message}`, { status: 500 });
    }
  },
};

// Utility function to extract hidden form values
function extractHiddenValue(html, fieldName) {
  const regex = new RegExp(`id="${fieldName}" value="(.*?)"`);
  const match = regex.exec(html);
  return match ? match[1] : null;
}
