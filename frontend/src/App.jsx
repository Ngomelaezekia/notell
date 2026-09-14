import router from "./utils/router";
import { RouterProvider } from "react-router-dom";
import { VideoFeedProvider } from "./context/VideoFeedContext";
import "./post-card-refinement.css";

function App() {
  return (
    <VideoFeedProvider>
      <RouterProvider router={router} />
    </VideoFeedProvider>
  );
}

export default App;
