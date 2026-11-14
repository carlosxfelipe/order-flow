// @ts-check

/**
 * @typedef {import("../dist/sdk").OrderSummary} OrderSummary
 * @typedef {import("../dist/sdk").OrderItem} OrderItem
 * @typedef {import("../dist/sdk").GetOrderResponse} GetOrderResponse
 * @typedef {import("../dist/sdk").CreateOrderRequest} CreateOrderRequest
 * @typedef {import("../dist/sdk").AddItemRequest} AddItemRequest
 * @typedef {import("../dist/sdk").PayOrderRequest} PayOrderRequest
 * @typedef {import("../dist/sdk").UploadReceiptRequest} UploadReceiptRequest
 * @typedef {import("../dist/sdk").User} User
 * @typedef {import("../dist/sdk").GotaClient} GotaClient
 * @typedef {import("../dist/sdk").GotaClientOptions} GotaClientOptions
 * @typedef {import("../dist/sdk").BaseRequestOptions} BaseRequestOptions
 * @typedef {import("../dist/sdk").BodyEncoding} BodyEncoding
 * @typedef {import("../dist/sdk").GotaRequestInit} GotaRequestInit
 * @typedef {import("../dist/sdk").GotaRequestConfig} GotaRequestConfig
 * @typedef {import("../dist/sdk").GotaResponseContext} GotaResponseContext
 */

const DEFAULT_RETRY_STATUSES = [408, 429, 500, 502, 503, 504];
const DEFAULT_RETRY_METHODS = ["GET", "HEAD", "OPTIONS", "PUT", "DELETE"];

/** @type {Record<string, BodyEncoding | undefined>} */
const BODY_ENCODINGS = {
  "POST /api/orders/:orderId/receipt": "multipart",
};

class GotaHttpError extends Error {
  /**
   * @param {number} status
   * @param {string} statusText
   * @param {any} data
   * @param {Headers} [headers]
   * @param {Response} [response]
   */
  constructor(status, statusText, data, headers, response) {
    super(errorMessage(data, statusText || "HTTP Error " + status));
    this.name = "GotaHttpError";
    this.status = status;
    this.statusText = statusText;
    this.data = data;
    this.headers = headers ?? new Headers();
    this.response = response;
    Object.setPrototypeOf(this, GotaHttpError.prototype);
  }
}

/**
 * @template [T=unknown]
 * @param {unknown} error
 * @returns {error is GotaHttpError}
 */
function isGotaHttpError(error) {
  return error instanceof GotaHttpError;
}

/**
 * @param {unknown[]} items
 * @returns {string}
 */
function formatValidationErrors(items) {
  const parts = items.map((item) => {
    if (typeof item === "string") return item;
    if (item && typeof item === "object") {
      const e = /** @type {Record<string, unknown>} */ (item);
      const loc = Array.isArray(e["loc"])
        ? e["loc"].map(String).join(".")
        : typeof e["field"] === "string"
          ? e["field"]
          : typeof e["path"] === "string"
            ? e["path"]
            : "";
      const msg =
        typeof e["msg"] === "string"
          ? e["msg"]
          : typeof e["message"] === "string"
            ? e["message"]
            : typeof e["error"] === "string"
              ? e["error"]
              : "";
      if (loc && msg) return loc + ": " + msg;
      if (msg) return msg;
    }
    return JSON.stringify(item);
  });
  return "Validation error: " + parts.join("; ");
}

/**
 * @param {unknown} data
 * @param {string} fallback
 * @returns {string}
 */
function errorMessage(data, fallback) {
  if (typeof data === "string" && data.trim()) return data.trim();
  if (data && typeof data === "object") {
    const d = /** @type {Record<string, unknown>} */ (data);
    if (typeof d["message"] === "string" && d["message"]) return d["message"];
    if (typeof d["error"] === "string" && d["error"]) return d["error"];
    if (typeof d["detail"] === "string" && d["detail"]) return d["detail"];
    if (Array.isArray(d["detail"]) && d["detail"].length > 0)
      return formatValidationErrors(d["detail"]);
    if (Array.isArray(d["errors"]) && d["errors"].length > 0)
      return formatValidationErrors(d["errors"]);
    if (typeof d["title"] === "string" && d["title"]) return d["title"];
  }
  return fallback;
}

/**
 * @param {unknown} value
 * @returns {value is Blob}
 */
function isBlob(value) {
  return (
    (typeof Blob !== "undefined" && value instanceof Blob) ||
    (typeof File !== "undefined" && value instanceof File)
  );
}

/**
 * @param {unknown} value
 * @returns {value is ReadableStream}
 */
function isReadableStream(value) {
  return (
    (typeof ReadableStream !== "undefined" &&
      value instanceof ReadableStream) ||
    (value !== null &&
      typeof value === "object" &&
      typeof (/** @type {{ getReader?: unknown }} */ (value).getReader) ===
        "function")
  );
}

/**
 * @param {unknown} body
 * @returns {body is BodyInit}
 */
function isBodyInit(body) {
  return (
    (typeof FormData !== "undefined" && body instanceof FormData) ||
    (typeof URLSearchParams !== "undefined" &&
      body instanceof URLSearchParams) ||
    isBlob(body) ||
    isReadableStream(body) ||
    (typeof ArrayBuffer !== "undefined" &&
      (body instanceof ArrayBuffer || ArrayBuffer.isView(body)))
  );
}

/**
 * Appends a query parameter value to URLSearchParams.
 * @param {URLSearchParams} search
 * @param {string} key
 * @param {unknown} value
 */
function appendQueryParam(search, key, value) {
  if (value === undefined || value === null) return;
  if (Array.isArray(value)) {
    for (const item of value) appendQueryParam(search, key, item);
  } else if (value instanceof Date) {
    search.append(key, value.toISOString());
  } else if (typeof value === "object") {
    search.append(key, JSON.stringify(value));
  } else {
    search.append(key, String(value));
  }
}

/**
 * @param {string} key
 * @param {unknown} value
 * @param {Array<[string, string | Blob]>} out
 */
function collectEntries(key, value, out) {
  if (value === undefined || value === null) return;
  if (Array.isArray(value)) {
    for (const item of value) collectEntries(key, item, out);
  } else if (isBlob(value)) {
    out.push([key, value]);
  } else if (value instanceof Date) {
    out.push([key, value.toISOString()]);
  } else if (typeof value === "object") {
    out.push([key, JSON.stringify(value)]);
  } else {
    out.push([key, String(value)]);
  }
}

/**
 * @param {object} source
 * @returns {Array<[string, string | Blob]>}
 */
function toEntries(source) {
  /** @type {Array<[string, string | Blob]>} */
  const out = [];
  for (const [key, value] of Object.entries(source))
    collectEntries(key, value, out);
  return out;
}

/**
 * @param {unknown} value
 * @param {Set<unknown>} [seen]
 * @returns {boolean}
 */
function hasBlobValue(value, seen = new Set()) {
  if (value === null || typeof value !== "object") return false;
  if (isBlob(value)) return true;
  if (seen.has(value)) return false;
  seen.add(value);
  if (Array.isArray(value))
    return value.some((item) => hasBlobValue(item, seen));
  if (value instanceof Date) return false;
  return Object.values(value).some((item) => hasBlobValue(item, seen));
}

/**
 * @param {string} endpoint
 * @param {string} template
 * @param {Record<string, unknown>} params
 * @returns {string}
 */
function resolvePath(endpoint, template, params) {
  return template
    .split("/")
    .map((segment) => {
      const kind = segment.charAt(0);
      if (kind !== ":" && kind !== "*") return segment;
      const name = segment.slice(1) || "wildcard";
      const value = params[name];
      if (value === undefined || value === null) {
        throw new Error(
          'Missing required path parameter "' + name + '" for ' + endpoint,
        );
      }
      if (kind === "*") {
        const parts = Array.isArray(value)
          ? value.map(String)
          : String(value).replace(/^\/+/, "").split("/");
        return parts.map(encodeURIComponent).join("/");
      }
      return encodeURIComponent(String(value));
    })
    .join("/");
}

/**
 * @param {string} endpoint
 * @param {unknown} body
 * @param {Headers} headers
 * @returns {BodyInit}
 */
function encodeBody(endpoint, body, headers) {
  if (typeof FormData !== "undefined" && body instanceof FormData) {
    headers.delete("Content-Type"); // let fetch set the multipart boundary
    return body;
  }
  if (isBodyInit(body)) return body;
  const encoding = BODY_ENCODINGS[endpoint] ?? "json";
  const contentType = headers.get("Content-Type") ?? "";
  if (body !== null && typeof body === "object") {
    if (
      encoding === "multipart" ||
      contentType.includes("multipart/form-data") ||
      hasBlobValue(body)
    ) {
      const form = new FormData();
      for (const [key, value] of toEntries(body)) form.append(key, value);
      headers.delete("Content-Type");
      return form;
    }
    if (
      encoding === "form" ||
      contentType.includes("application/x-www-form-urlencoded")
    ) {
      const form = new URLSearchParams();
      for (const [key, value] of toEntries(body)) {
        if (typeof value === "string") form.append(key, value);
      }
      return form;
    }
  }
  if (
    typeof body === "string" &&
    (encoding === "raw" ||
      (contentType !== "" && !contentType.includes("json")))
  ) {
    return body;
  }
  if (!contentType) headers.set("Content-Type", "application/json");
  return JSON.stringify(body);
}

/**
 * @param {number | undefined} timeoutMs
 * @returns {AbortSignal | undefined}
 */
function createTimeoutSignal(timeoutMs) {
  if (timeoutMs === undefined || timeoutMs <= 0) return undefined;
  if (
    typeof AbortSignal !== "undefined" &&
    typeof AbortSignal.timeout === "function"
  ) {
    return AbortSignal.timeout(timeoutMs);
  }
  if (typeof AbortController === "undefined") return undefined;
  const controller = new AbortController();
  const timeoutReason =
    typeof DOMException !== "undefined"
      ? new DOMException("The operation timed out.", "TimeoutError")
      : new Error("The operation timed out.");
  const timer = setTimeout(() => controller.abort(timeoutReason), timeoutMs);
  if (
    timer &&
    typeof timer === "object" &&
    "unref" in timer &&
    typeof (/** @type {{ unref?: unknown }} */ (timer).unref) === "function"
  ) {
    /** @type {{ unref: () => void }} */ (timer).unref();
  }
  return controller.signal;
}

/**
 * @param {AbortSignal | undefined} a
 * @param {AbortSignal | undefined} b
 * @returns {AbortSignal | undefined}
 */
function combineSignals(a, b) {
  if (!a) return b;
  if (!b) return a;
  const any = /** @type {any} */ (AbortSignal).any;
  if (typeof any === "function") return any.call(AbortSignal, [a, b]);
  const controller = new AbortController();
  const cleanup = () => {
    a.removeEventListener("abort", onAbortA);
    b.removeEventListener("abort", onAbortB);
  };
  const onAbortA = () => {
    cleanup();
    controller.abort(a.reason);
  };
  const onAbortB = () => {
    cleanup();
    controller.abort(b.reason);
  };
  if (a.aborted) controller.abort(a.reason);
  else if (b.aborted) controller.abort(b.reason);
  else {
    a.addEventListener("abort", onAbortA, { once: true });
    b.addEventListener("abort", onAbortB, { once: true });
  }
  return controller.signal;
}

/**
 * @param {number} baseDelay
 * @param {number} attempt
 * @param {Response} [response]
 * @returns {number}
 */
function retryDelay(baseDelay, attempt, response) {
  let delay =
    baseDelay * Math.pow(2, attempt - 1) * (0.5 + Math.random() * 0.5);
  const retryAfter = response?.headers?.get("Retry-After");
  if (retryAfter) {
    const seconds = parseInt(retryAfter, 10);
    if (!isNaN(seconds)) {
      delay = seconds * 1000;
    } else {
      const dateMs = Date.parse(retryAfter);
      if (!isNaN(dateMs)) delay = Math.max(0, dateMs - Date.now());
    }
  }
  return delay;
}

/**
 * @param {number} ms
 * @returns {Promise<void>}
 */
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * @param {Response} res
 * @returns {Promise<unknown>}
 */
async function readErrorBody(res) {
  if (typeof res.text === "function") {
    const text = await res.text().catch(() => "");
    try {
      return text ? JSON.parse(text) : { message: res.statusText };
    } catch {
      return { message: text || res.statusText };
    }
  }
  if (typeof res.json === "function") {
    return res.json().catch(() => ({ message: res.statusText }));
  }
  return { message: res.statusText };
}

/**
 * @param {Response} res
 * @returns {Promise<any>}
 */
async function readBody(res) {
  if (
    res.status === 204 ||
    res.status === 205 ||
    res.headers.get("content-length") === "0"
  )
    return undefined;
  const contentType = res.headers.get("content-type") ?? "";
  if (
    contentType.includes("application/json") ||
    contentType.includes("+json")
  ) {
    if (typeof res.text === "function") {
      const text = await res.text();
      return text ? JSON.parse(text) : undefined;
    }
    return res.json();
  }
  if (contentType.includes("text/event-stream") && res.body) {
    return typeof TextDecoderStream !== "undefined"
      ? res.body.pipeThrough(new TextDecoderStream())
      : res.body;
  }
  const isBinary =
    /^(image|audio|video)\/|application\/(octet-stream|pdf|zip|gzip|vnd\.|x-)/.test(
      contentType,
    );
  if (isBinary && typeof res.blob === "function") return res.blob();
  if (contentType.startsWith("text/") || contentType.includes("xml"))
    return res.text();
  if (typeof res.blob === "function") return res.blob();
  return res.text();
}

/**
 * Exact runtime implementation of createGotaClient() generated in dist/sdk.ts
 * @param {string} [baseUrl=""]
 * @param {GotaClientOptions} [clientOptions={}]
 * @returns {GotaClient}
 */
function createGotaClient(baseUrl = "", clientOptions = {}) {
  const root = baseUrl.replace(/\/+$/, "");

  /**
   * @param {string} endpoint
   * @param {any} options
   * @returns {Promise<any>}
   */
  async function request(endpoint, options) {
    const separator = endpoint.indexOf(" ");
    const method = endpoint.slice(0, separator);
    const template = endpoint.slice(separator + 1);
    const path = resolvePath(endpoint, template, options.params ?? {});

    let url = root + (path.startsWith("/") ? path : "/" + path);
    if (options.query) {
      const search = new URLSearchParams();
      for (const [key, value] of Object.entries(options.query))
        appendQueryParam(search, key, value);
      const qs = search.toString();
      if (qs) url += (url.includes("?") ? "&" : "?") + qs;
    }

    const baseHeaders =
      typeof clientOptions.headers === "function"
        ? await clientOptions.headers()
        : clientOptions.headers;
    const headers = new Headers(baseHeaders);
    if (options.headers) {
      for (const [key, value] of Object.entries(options.headers))
        headers.set(key, String(value));
    }

    /** @type {GotaRequestInit} */
    const init = { method, headers };
    if (options.body !== undefined) {
      init.body = encodeBody(endpoint, options.body, headers);
    }
    const credentials = options.credentials ?? clientOptions.credentials;
    if (credentials) init.credentials = credentials;
    const cache = options.cache ?? clientOptions.cache;
    if (cache) init.cache = cache;
    const nextConfig =
      clientOptions.next || options.next
        ? { ...clientOptions.next, ...options.next }
        : undefined;
    if (nextConfig) init.next = nextConfig;
    const duplex =
      options.duplex ?? (isReadableStream(init.body) ? "half" : undefined);
    if (duplex) init.duplex = duplex;

    const signal = combineSignals(
      options.signal,
      createTimeoutSignal(options.timeoutMs ?? clientOptions.timeoutMs),
    );
    if (signal) init.signal = signal;

    if (clientOptions.onRequest)
      await clientOptions.onRequest({ endpoint, url, init });
    if (options.onRequest) await options.onRequest({ endpoint, url, init });

    const retryOpt = options.retry ?? clientOptions.retry;
    let maxAttempts = 1;
    let baseDelay = 300;
    let retryStatuses = DEFAULT_RETRY_STATUSES;
    let retryMethods = DEFAULT_RETRY_METHODS;
    if (typeof retryOpt === "number" && retryOpt > 0) {
      maxAttempts = 1 + retryOpt;
    } else if (typeof retryOpt === "object" && retryOpt !== null) {
      if (typeof retryOpt.attempts === "number" && retryOpt.attempts > 0)
        maxAttempts = 1 + retryOpt.attempts;
      if (typeof retryOpt.delayMs === "number" && retryOpt.delayMs >= 0)
        baseDelay = retryOpt.delayMs;
      if (Array.isArray(retryOpt.statusCodes))
        retryStatuses = retryOpt.statusCodes;
      if (Array.isArray(retryOpt.methods))
        retryMethods = retryOpt.methods.map((/** @type {string} */ m) =>
          m.toUpperCase(),
        );
    }
    if (!retryMethods.includes(method.toUpperCase())) maxAttempts = 1;

    const fetchFn = clientOptions.fetch ?? globalThis.fetch.bind(globalThis);
    /** @type {Response} */
    let res;
    let attempt = 0;
    while (true) {
      try {
        res = await fetchFn(url, init);
      } catch (fetchErr) {
        if (attempt < maxAttempts - 1 && !init.signal?.aborted) {
          attempt++;
          await sleep(retryDelay(baseDelay, attempt));
          continue;
        }
        const errObj =
          fetchErr instanceof Error ? fetchErr : new Error(String(fetchErr));
        if (clientOptions.onError)
          await clientOptions.onError(errObj, { endpoint, url });
        if (options.onError) await options.onError(errObj, { endpoint, url });
        throw fetchErr;
      }
      if (
        !res.ok &&
        attempt < maxAttempts - 1 &&
        retryStatuses.includes(res.status) &&
        !init.signal?.aborted
      ) {
        attempt++;
        await sleep(retryDelay(baseDelay, attempt, res));
        continue;
      }
      break;
    }

    /** @type {GotaResponseContext} */
    const context = { endpoint, url };
    if (clientOptions.onResponse) await clientOptions.onResponse(res, context);
    if (options.onResponse) await options.onResponse(res, context);

    if (!res.ok) {
      const data = await readErrorBody(res);
      const httpErr = new GotaHttpError(
        res.status,
        res.statusText,
        data,
        res.headers,
        res,
      );
      if (clientOptions.onError) await clientOptions.onError(httpErr, context);
      if (options.onError) await options.onError(httpErr, context);
      throw httpErr;
    }
    return readBody(res);
  }

  return {
    baseUrl: root,
    fetch(endpoint, ...args) {
      const [options] = /** @type {[any?]} */ (args);
      return request(String(endpoint), options ?? {});
    },
  };
}

// @ts-ignore - Vue is loaded globally via CDN
const { createApp, ref, reactive, computed, onMounted } = Vue;

createApp({
  setup() {
    // ----------------------------------------------------
    // Server & Connection State
    // ----------------------------------------------------
    const baseUrl = ref(
      localStorage.getItem("orderflow_baseUrl") || "http://localhost:3000",
    );
    const docsUrl = computed(
      () => (baseUrl.value ? baseUrl.value.replace(/\/+$/, "") : "") + "/docs",
    );
    const connectionError = ref(false);
    const notification = ref(
      /** @type {{ text: string; type: 'success' | 'error' | 'info' } | null} */ (
        null
      ),
    );

    // ----------------------------------------------------
    // Auth & User State
    // ----------------------------------------------------
    const token = ref(localStorage.getItem("orderflow_jwt") || "");
    const currentUser = ref(/** @type {User | null} */ (null));
    const isAuthenticating = ref(false);
    const showAuthModal = ref(false);
    const authTab = ref(/** @type {'login' | 'register'} */ ("login"));
    const authError = ref(/** @type {string | null} */ (null));

    const loginForm = reactive({
      username: "carlos",
      password: "carlos123",
    });

    const registerForm = reactive({
      username: "",
      password: "",
    });

    // Seed users quick shortcuts
    const seedAccounts = [
      {
        username: "admin",
        label: "Admin (All Access)",
        role: "Admin",
        pass: "admin123",
      },
      {
        username: "carlos",
        label: "Carlos (Customer)",
        role: "Customer",
        pass: "carlos123",
      },
      {
        username: "maria",
        label: "Maria (Customer)",
        role: "Customer",
        pass: "maria123",
      },
      {
        username: "luis",
        label: "Luis (Customer)",
        role: "Customer",
        pass: "luis123",
      },
    ];

    // ----------------------------------------------------
    // Orders State
    // ----------------------------------------------------
    const orders = ref(/** @type {OrderSummary[]} */ ([]));
    const loadingOrders = ref(false);
    const selectedStatusFilter = ref("all"); // 'all' | 'Created' | 'Paid' | 'Shipped' | 'Cancelled'

    // Selected / Active Order Detail
    const selectedOrderId = ref("");
    const selectedOrder = ref(/** @type {GetOrderResponse | null} */ (null));
    const loadingOrderDetail = ref(false);
    const lookupError = ref(/** @type {string | null} */ (null));

    // Create Order Form
    const creatingOrder = ref(false);
    const createOrderError = ref(
      /** @type {{ message: string; fields?: Record<string, string> | null } | null} */ (
        null
      ),
    );

    /**
     * Normalizes validation error arrays into clean comma-separated strings for UI display
     * @param {Record<string, string[] | string>} errors
     * @returns {Record<string, string>}
     */
    function formatErrors(errors) {
      /** @type {Record<string, string>} */
      const result = {};
      for (const [key, val] of Object.entries(errors)) {
        result[key] = Array.isArray(val) ? val.join(", ") : String(val);
      }
      return result;
    }
    const showShippingFields = ref(false);
    const newOrderForm = reactive({
      customerName: "",
      shippingAddress: {
        street: "Av. Paulista, 1000",
        city: "São Paulo",
        zipCode: "01310100",
      },
    });

    // Add Item Form
    const addingItem = ref(false);
    const addItemError = ref(/** @type {string | null} */ (null));
    const newItemForm = reactive({
      productName: "Coca-Cola Can 350ml",
      price: 5.5,
      quantity: 1,
    });

    // Payment Form
    const payingOrder = ref(false);
    const paymentForm = reactive({
      method: /** @type {'pix' | 'credit_card' | 'boleto'} */ ("pix"),
      currency: /** @type {'BRL' | 'USD' | 'EUR'} */ ("BRL"),
    });

    // Lifecycle Actions (Ship, Cancel)
    const actionLoading = ref(false);

    // Receipt Upload Form
    const uploadingReceipt = ref(false);
    const receiptNotes = ref("");
    const receiptFile = ref(/** @type {File | null} */ (null));
    const receiptSuccess = ref(/** @type {string | null} */ (null));
    const receiptFileInput = ref(/** @type {HTMLInputElement | null} */ (null));

    // SSE Tracking
    const isTracking = ref(false);
    const trackingEvents = ref(
      /** @type {{ id: number; time: string; event: string; data: any }[]} */ ([]),
    );

    // Raw JSON Inspector Modal
    const inspectorPayload = ref(/** @type {any} */ (null));
    const inspectorTitle = ref("API Response Inspector");

    // ----------------------------------------------------
    // API Client Factory
    // ----------------------------------------------------
    const client = () => {
      /** @type {Record<string, string>} */
      const headers = {};
      if (token.value) {
        headers["Authorization"] = `Bearer ${token.value}`;
      }
      return createGotaClient(baseUrl.value, { headers });
    };

    /**
     * @param {string} text
     * @param {'success' | 'error' | 'info'} [type="success"]
     */
    function notify(text, type = "success") {
      notification.value = { text, type };
      setTimeout(() => {
        if (notification.value?.text === text) {
          notification.value = null;
        }
      }, 4000);
    }

    // ----------------------------------------------------
    // Auth Methods
    // ----------------------------------------------------
    /**
     * @param {string} username
     * @param {string} password
     */
    async function quickLogin(username, password) {
      loginForm.username = username;
      loginForm.password = password;
      await handleLogin();
    }

    async function handleLogin() {
      isAuthenticating.value = true;
      authError.value = null;
      try {
        const res = await client().fetch("POST /api/auth/login", {
          body: {
            username: loginForm.username,
            password: loginForm.password,
          },
        });
        token.value = res.token;
        localStorage.setItem("orderflow_jwt", res.token);
        await fetchCurrentUser();
        showAuthModal.value = false;
        notify(
          `Welcome back, ${currentUser.value?.username || loginForm.username}!`,
        );
        await loadOrders();
      } catch (/** @type {any} */ err) {
        authError.value = err.message || "Failed to authenticate";
      } finally {
        isAuthenticating.value = false;
      }
    }

    async function handleRegister() {
      isAuthenticating.value = true;
      authError.value = null;
      try {
        const res = await client().fetch("POST /api/auth/register", {
          body: {
            username: registerForm.username,
            password: registerForm.password,
          },
        });
        token.value = res.token;
        localStorage.setItem("orderflow_jwt", res.token);
        currentUser.value = {
          id: res.id,
          username: res.username,
          role: res.role,
        };
        showAuthModal.value = false;
        notify(`Account created successfully! Logged in as ${res.username}.`);
        await loadOrders();
      } catch (/** @type {any} */ err) {
        authError.value = err.message || "Registration failed";
      } finally {
        isAuthenticating.value = false;
      }
    }

    async function fetchCurrentUser() {
      if (!token.value) {
        currentUser.value = null;
        return;
      }
      try {
        currentUser.value = await client().fetch("GET /api/auth/me");
        if (currentUser.value && !newOrderForm.customerName) {
          newOrderForm.customerName = currentUser.value.username;
        }
      } catch {
        // Token might be invalid or expired
        logout();
      }
    }

    function logout() {
      token.value = "";
      currentUser.value = null;
      localStorage.removeItem("orderflow_jwt");
      orders.value = [];
      selectedOrder.value = null;
      selectedOrderId.value = "";
      notify("Logged out successfully.", "info");
    }

    async function deleteAccount() {
      if (
        !confirm(
          "Are you sure you want to permanently delete your account? This action cannot be undone.",
        )
      ) {
        return;
      }
      try {
        await client().fetch("DELETE /api/auth/account");
        notify("Account deleted permanently.", "info");
        logout();
      } catch (/** @type {any} */ err) {
        notify(err.message || "Failed to delete account", "error");
      }
    }

    // ----------------------------------------------------
    // Orders Methods
    // ----------------------------------------------------
    async function loadOrders() {
      if (!token.value) return;
      loadingOrders.value = true;
      connectionError.value = false;
      try {
        const res = await client().fetch("GET /api/orders");
        orders.value = res.orders || [];
        // If an order is currently selected, refresh its details too
        if (selectedOrderId.value) {
          fetchOrderById(selectedOrderId.value, false);
        }
      } catch (/** @type {any} */ err) {
        if (err.status === 401) {
          logout();
        } else {
          connectionError.value = true;
        }
      } finally {
        loadingOrders.value = false;
      }
    }

    /**
     * @param {string} id
     * @param {boolean} [openInspector=false]
     */
    async function fetchOrderById(id, openInspector = false) {
      if (!id) return;
      loadingOrderDetail.value = true;
      lookupError.value = null;
      try {
        const res = await client().fetch("GET /api/orders/:orderId", {
          params: { orderId: id.trim() },
        });
        selectedOrder.value = res;
        selectedOrderId.value = res.id;
        if (openInspector) {
          inspectPayload(res, `Order: ${res.id}`);
        }
      } catch (/** @type {any} */ err) {
        if (err instanceof GotaHttpError) {
          lookupError.value = `${err.status}: ${err.message}`;
        } else {
          lookupError.value = err?.message || String(err);
        }
      } finally {
        loadingOrderDetail.value = false;
      }
    }

    async function createOrder() {
      creatingOrder.value = true;
      createOrderError.value = null;
      try {
        /** @type {CreateOrderRequest} */
        const payload = {
          customerName: newOrderForm.customerName.trim(),
        };
        if (showShippingFields.value && newOrderForm.shippingAddress.street) {
          payload.shippingAddress = {
            street: newOrderForm.shippingAddress.street,
            city: newOrderForm.shippingAddress.city,
            zipCode: newOrderForm.shippingAddress.zipCode,
          };
        }

        const res = await client().fetch("POST /api/orders", {
          body: payload,
        });

        notify(`Order created! #${res.orderId}`);
        await loadOrders();
        await fetchOrderById(res.orderId);
      } catch (/** @type {any} */ err) {
        if (err instanceof GotaHttpError) {
          const data = err.data;
          createOrderError.value = {
            message: err.message,
            fields:
              data && "errors" in data && data.errors
                ? formatErrors(data.errors)
                : null,
          };
        } else {
          createOrderError.value = {
            message: err?.message || String(err),
            fields: null,
          };
        }
      } finally {
        creatingOrder.value = false;
      }
    }

    // ----------------------------------------------------
    // Order Lifecycle Actions
    // ----------------------------------------------------
    async function addItemToOrder() {
      if (!selectedOrder.value) return;
      addingItem.value = true;
      addItemError.value = null;
      try {
        await client().fetch("POST /api/orders/:orderId/items", {
          params: { orderId: selectedOrder.value.id },
          body: {
            productName: newItemForm.productName,
            price: Number(newItemForm.price),
            quantity: Number(newItemForm.quantity),
          },
        });
        notify(`Added "${newItemForm.productName}" to order.`);
        newItemForm.productName = "";
        newItemForm.price = 10.0;
        newItemForm.quantity = 1;
        await fetchOrderById(selectedOrder.value.id);
        await loadOrders();
      } catch (/** @type {any} */ err) {
        addItemError.value = err.message;
      } finally {
        addingItem.value = false;
      }
    }

    /**
     * @param {string} itemId
     */
    async function removeItem(itemId) {
      if (!selectedOrder.value) return;
      try {
        await client().fetch("DELETE /api/orders/:orderId/items/:itemId", {
          params: {
            orderId: selectedOrder.value.id,
            itemId: itemId,
          },
        });
        notify("Item removed from order.");
        await fetchOrderById(selectedOrder.value.id);
        await loadOrders();
      } catch (/** @type {any} */ err) {
        notify(err.message, "error");
      }
    }

    async function payOrder() {
      if (!selectedOrder.value) return;
      actionLoading.value = true;
      try {
        const res = await client().fetch("POST /api/orders/:orderId/pay", {
          params: { orderId: selectedOrder.value.id },
          body: {
            method: paymentForm.method,
            currency: paymentForm.currency,
          },
        });
        notify(res.message || "Order paid successfully!");
        await fetchOrderById(selectedOrder.value.id);
        await loadOrders();
      } catch (/** @type {any} */ err) {
        notify(err.message, "error");
      } finally {
        actionLoading.value = false;
      }
    }

    async function shipOrder() {
      if (!selectedOrder.value) return;
      actionLoading.value = true;
      try {
        const res = await client().fetch("POST /api/orders/:orderId/ship", {
          params: { orderId: selectedOrder.value.id },
        });
        notify(res.message || "Order shipped successfully!");
        await fetchOrderById(selectedOrder.value.id);
        await loadOrders();
      } catch (/** @type {any} */ err) {
        notify(err.message, "error");
      } finally {
        actionLoading.value = false;
      }
    }

    async function cancelOrder() {
      if (!selectedOrder.value) return;
      if (!confirm("Are you sure you want to cancel this order?")) return;
      actionLoading.value = true;
      try {
        const res = await client().fetch("POST /api/orders/:orderId/cancel", {
          params: { orderId: selectedOrder.value.id },
        });
        notify(res.message || "Order cancelled.");
        await fetchOrderById(selectedOrder.value.id);
        await loadOrders();
      } catch (/** @type {any} */ err) {
        notify(err.message, "error");
      } finally {
        actionLoading.value = false;
      }
    }

    // ----------------------------------------------------
    // Receipt File Upload (Multipart Form Data)
    // ----------------------------------------------------
    /**
     * @param {any} event
     */
    function handleFileSelected(event) {
      const file = event.target.files?.[0] || null;
      receiptFile.value = file;
      receiptSuccess.value = null;
    }

    async function uploadReceipt() {
      if (!selectedOrder.value || !receiptFile.value) return;
      uploadingReceipt.value = true;
      receiptSuccess.value = null;
      try {
        /** @type {UploadReceiptRequest} */
        const payload = {
          receipt: receiptFile.value,
        };
        if (receiptNotes.value.trim()) {
          payload.notes = receiptNotes.value.trim();
        }

        const res = await client().fetch("POST /api/orders/:orderId/receipt", {
          params: { orderId: selectedOrder.value.id },
          body: payload,
        });

        receiptSuccess.value = `${res.message} (${res.filename}, ${(res.size / 1024).toFixed(1)} KB)`;
        notify("Receipt uploaded successfully!");
        receiptFile.value = null;
        receiptNotes.value = "";
        if (receiptFileInput.value) receiptFileInput.value.value = "";
        await fetchOrderById(selectedOrder.value.id);
      } catch (/** @type {any} */ err) {
        notify(err.message || "Failed to upload receipt", "error");
      } finally {
        uploadingReceipt.value = false;
      }
    }

    // ----------------------------------------------------
    // Server-Sent Events (SSE) Live Tracking
    // ----------------------------------------------------
    async function startSSETracking() {
      if (!selectedOrder.value) return;
      isTracking.value = true;
      trackingEvents.value = [];

      try {
        const stream = await client().fetch("GET /api/orders/:orderId/track", {
          params: { orderId: selectedOrder.value.id },
        });

        const reader = stream.getReader();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;
          buffer += value;

          // Parse event-stream format chunks
          const lines = buffer.split("\n");
          buffer = lines.pop() || "";

          let currentEvent = "message";
          for (const line of lines) {
            const trimmed = line.trim();
            if (trimmed.startsWith("event:")) {
              currentEvent = trimmed.slice(6).trim();
            } else if (trimmed.startsWith("data:")) {
              const rawData = trimmed.slice(5).trim();
              let parsedData = rawData;
              try {
                parsedData = JSON.parse(rawData);
              } catch {}

              trackingEvents.value.push({
                id: Date.now() + Math.random(),
                time: new Date().toLocaleTimeString(),
                event: currentEvent,
                data: parsedData,
              });
              currentEvent = "message";
            }
          }
        }

        notify("Live tracking stream ended.");
      } catch (/** @type {any} */ err) {
        trackingEvents.value.push({
          id: Date.now(),
          time: new Date().toLocaleTimeString(),
          event: "error",
          data: err.message || "Tracking failed",
        });
      } finally {
        isTracking.value = false;
      }
    }

    // ----------------------------------------------------
    // UI Helpers & Badges
    // ----------------------------------------------------
    const filteredOrders = computed(() => {
      if (selectedStatusFilter.value === "all") return orders.value;
      return orders.value.filter(
        (/** @type {OrderSummary} */ o) =>
          o.status.toLowerCase() === selectedStatusFilter.value.toLowerCase(),
      );
    });

    /**
     * @param {string} status
     */
    function statusBadgeClass(status) {
      switch (status) {
        case "Created":
          return "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/25";
        case "Paid":
          return "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/25";
        case "Shipped":
          return "bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border-indigo-500/25";
        case "Cancelled":
          return "bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/25";
        default:
          return "bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/25";
      }
    }

    /**
     * @param {number | string} amount
     */
    function formatCurrency(amount) {
      return Number(amount || 0).toLocaleString("pt-BR", {
        style: "currency",
        currency: "BRL",
      });
    }

    /**
     * @param {any} payload
     * @param {string} [title="API Response"]
     */
    function inspectPayload(payload, title = "API Response") {
      inspectorPayload.value = payload;
      inspectorTitle.value = title;
    }

    function updateBaseUrl() {
      localStorage.setItem("orderflow_baseUrl", baseUrl.value);
      loadOrders();
    }

    // ----------------------------------------------------
    // Init Lifecycle
    // ----------------------------------------------------
    onMounted(async () => {
      if (token.value) {
        await fetchCurrentUser();
        await loadOrders();
      } else {
        // Auto-login with carlos by default for frictionless local showcase
        await quickLogin("carlos", "carlos123");
      }
    });

    return {
      baseUrl,
      docsUrl,
      connectionError,
      notification,
      token,
      currentUser,
      isAuthenticating,
      showAuthModal,
      authTab,
      authError,
      loginForm,
      registerForm,
      seedAccounts,
      orders,
      filteredOrders,
      loadingOrders,
      selectedStatusFilter,
      selectedOrderId,
      selectedOrder,
      loadingOrderDetail,
      lookupError,
      creatingOrder,
      createOrderError,
      showShippingFields,
      newOrderForm,
      addingItem,
      addItemError,
      newItemForm,
      payingOrder,
      paymentForm,
      actionLoading,
      uploadingReceipt,
      receiptNotes,
      receiptFile,
      receiptSuccess,
      receiptFileInput,
      isTracking,
      trackingEvents,
      inspectorPayload,
      inspectorTitle,
      quickLogin,
      handleLogin,
      handleRegister,
      logout,
      deleteAccount,
      loadOrders,
      fetchOrderById,
      createOrder,
      addItemToOrder,
      removeItem,
      payOrder,
      shipOrder,
      cancelOrder,
      handleFileSelected,
      uploadReceipt,
      startSSETracking,
      statusBadgeClass,
      formatCurrency,
      inspectPayload,
      updateBaseUrl,
    };
  },
}).mount("#app");
