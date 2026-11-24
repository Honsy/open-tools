import AdminAccess from "./pages/admin/Access.vue";
import AdminLayout from "./pages/admin/Layout.vue";
import AdminReview from "./pages/admin/Review.vue";
import AdminLinks from "./pages/admin/Links.vue";
import AdminCategories from "./pages/admin/Categories.vue";
import AdminArticles from "./pages/admin/Articles.vue";
import AdminTags from "./pages/admin/Tags.vue";
import AdminFeedback from "./pages/admin/Feedback.vue";
import AdminUsers from "./pages/admin/Users.vue";
import AdminInternetRank from "./pages/admin/InternetRank.vue";

const adminChildren = [
  { path: "", component: AdminReview },
  { path: "links", component: AdminLinks },
  { path: "categories", component: AdminCategories },
  { path: "articles", component: AdminArticles },
  { path: "tags", component: AdminTags },
  { path: "feedback", component: AdminFeedback },
  { path: "users", component: AdminUsers },
  { path: "internet-rank", component: AdminInternetRank },
];

export const routes = [
  { path: "/login", redirect: "/" },
  { path: "/access", component: AdminAccess },
  { path: "/", component: AdminLayout, children: adminChildren },
];
