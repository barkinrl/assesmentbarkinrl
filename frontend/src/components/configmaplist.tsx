import React, { useEffect, useState } from "react";
import { useAuth } from "react-oidc-context";

type ConfigMap = {
  name: string;
  data?: Record<string, string>;
};

const ConfigMapList: React.FC = () => {
  const [configMaps, setConfigMaps] = useState<ConfigMap[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<string | null>(null);
  const [editValue, setEditValue] = useState<string>("");

  const auth = useAuth();

  // Fetch configmaps
  useEffect(() => {
    if (!auth.user?.access_token) return;

    fetch("/api", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + auth.user.access_token,
      },
      body: JSON.stringify({ action: "get_configmaps" }),
    })
      .then((res) => res.json())
      .then((data) => {
        setConfigMaps(data);
        setLoading(false);
      });
  }, [auth.user]);

  // Fetch single configmap data for editing
  const handleEdit = (name: string) => {
    if (!auth.user?.access_token) return;
    fetch("/api", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + auth.user.access_token,
      },
      body: JSON.stringify({ action: "get_configmap", name }),
    })
      .then((res) => res.json())
      .then((data) => {
        setEditing(name);
        setEditValue(JSON.stringify(data.data, null, 2));
      });
  };

  const handleSave = (name: string) => {
    if (!auth.user?.access_token) return;
    fetch("/api", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + auth.user.access_token,
      },
      body: JSON.stringify({
        action: "update_configmap",
        name,
        data: JSON.parse(editValue),
      }),
    }).then(() => {
      setEditing(null);
      // Optionally, refresh the list
    });
  };

  const handleDelete = (name: string) => {
    if (!auth.user?.access_token) return;
    fetch("/api", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + auth.user.access_token,
      },
      body: JSON.stringify({
        action: "delete_configmap",
        name,
      }),
    }).then(() => setConfigMaps(configMaps.filter((c) => c.name !== name)));
  };

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      <h2>ConfigMaps</h2>
      <ul>
        {configMaps.map((cm) => (
          <li key={cm.name}>
            {cm.name}
            <button
              onClick={() => handleEdit(cm.name)}
              style={{ marginLeft: "10px" }}
            >
              Edit
            </button>
            <button
              onClick={() => handleDelete(cm.name)}
              style={{ marginLeft: "10px" }}
            >
              Delete
            </button>
            {editing === cm.name && (
              <div>
                <textarea
                  rows={8}
                  cols={40}
                  value={editValue}
                  onChange={(e) => setEditValue(e.target.value)}
                  style={{ display: "block", marginTop: "10px" }}
                />
                <button
                  onClick={() => handleSave(cm.name)}
                  style={{ marginTop: "5px" }}
                >
                  Save
                </button>
                <button
                  onClick={() => setEditing(null)}
                  style={{ marginLeft: "5px" }}
                >
                  Cancel
                </button>
              </div>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default ConfigMapList;
