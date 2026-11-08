// @ts-check

/**
 * @typedef {import("../dist/sdk").OrderSummary} OrderSummary
 * @typedef {import("../dist/sdk").OrderItem} OrderItem
 * @typedef {import("../dist/sdk").GetOrderResponse} GetOrderResponse
 * @typedef {import("../dist/sdk").CreateOrderRequest} CreateOrderRequest
 * @typedef {import("../dist/sdk").AddItemRequest} AddItemRequest
 * @typedef {import("../dist/sdk").PayOrderRequest} PayOrderRequest
 * @typedef {import("../dist/sdk").User} User
 * @typedef {import("../dist/sdk").GotaClient} GotaClient
 */

/**
 * Exact runtime implementation of createGotaClient() generated in dist/sdk.ts
 * @param {string} [baseUrl=""]
 * @param {import("../dist/sdk").GotaClientOptions} [clientOptions={}]
 * @returns {GotaClient}
 */
function createGotaClient(baseUrl = "", clientOptions = {}) {
  return {
    baseUrl,
    async fetch(endpoint, ...args) {
      const options = /** @type {any} */ (args[0] || {});
      const [method, pathTemplate] = endpoint.split(" ");
      let path = pathTemplate;
      if (options.params) {
        const keys = Object.keys(options.params).sort(
          (a, b) => b.length - a.length,
        );
        for (const k of keys) {
          const v = options.params[k];
          path = path.replace(":" + k, encodeURIComponent(String(v)));
          path = path.replace("*" + k, String(v));
        }
      }
      let url = baseUrl + path;
      if (options.query) {
        const q = new URLSearchParams();
        for (const [k, v] of Object.entries(options.query)) {
          if (v !== undefined && v !== null) {
            q.set(k, String(v));
          }
        }
        const qs = q.toString();
        if (qs) url += (url.includes("?") ? "&" : "?") + qs;
      }
      const baseHeaders =
        typeof clientOptions.headers === "function"
          ? await clientOptions.headers()
          : clientOptions.headers;
      const headers = new Headers(baseHeaders);
      if (options.headers) {
        for (const [k, v] of Object.entries(options.headers)) {
          headers.set(k, String(v));
        }
      }
      let body = undefined;
      if (options.body !== undefined) {
        if (
          typeof FormData !== "undefined" &&
          options.body instanceof FormData
        ) {
          body = options.body;
        } else {
          if (!headers.has("Content-Type")) {
            headers.set("Content-Type", "application/json");
          }
          body = JSON.stringify(options.body);
        }
      }
      const fetchFn = clientOptions.fetch || fetch;
      const res = await fetchFn(url, {
        method,
        headers,
        body,
        credentials: options.credentials ?? clientOptions.credentials,
        signal: options.signal,
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({ message: res.statusText }));
        throw Object.assign(
          new Error((err && (err.message || err.error)) || res.statusText),
          { status: res.status, data: err },
        );
      }
      if (res.status === 204) return /** @type {any} */ (undefined);
      const ctype = res.headers.get("content-type") || "";
      if (ctype.includes("application/json")) {
        return res.json();
      }
      if (ctype.includes("text/event-stream") && res.body) {
        return /** @type {any} */ (
          typeof TextDecoderStream !== "undefined"
            ? res.body.pipeThrough(new TextDecoderStream())
            : res.body
        );
      }
      return /** @type {any} */ (res.text());
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
        lookupError.value = `${err.status || "Error"}: ${err.message}`;
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
        createOrderError.value = {
          message: err.message,
          fields: err.data?.errors || null,
        };
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
        const formData = new FormData();
        formData.append("receipt", receiptFile.value);
        if (receiptNotes.value.trim()) {
          formData.append("notes", receiptNotes.value.trim());
        }

        const res = await client().fetch("POST /api/orders/:orderId/receipt", {
          params: { orderId: selectedOrder.value.id },
          body: formData,
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
