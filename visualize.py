import json
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import plotly.express as px
import streamlit as st
from datetime import datetime

# Load data
with open("result.json", "r") as f:
    data = json.load(f)

df = pd.json_normalize(data)
df["timestamp"] = pd.to_datetime(df["timestamp"])
df["duration_ms"] = df["duration"].str.replace("ms", "").astype(float)
df["second"] = df["timestamp"].dt.floor("s")
df = df[df["node_id"].notnull()]  # Remove missing node_id

# Layout setup
st.set_page_config(layout="centered")
st.title("📊 Cassandra Benchmarking Results")

# Plot settings
sns.set(style="whitegrid")
plt.rcParams.update({'axes.titlesize': 10, 'axes.labelsize': 8})

def compact_fig():
    return plt.subplots(figsize=(5, 2.5))


st.subheader("1️⃣ Node-wise Throughput Over Time")
tp = df.groupby(["second", "node_id", "action"]).size().reset_index(name="count")
tp["label"] = "Node " + tp["node_id"].astype(str) + " - " + tp["action"]
fig_tp = px.line(tp, x="second", y="count", color="label", markers=True)
fig_tp.update_layout(height=350, title="Throughput (ops/sec)", margin=dict(l=20, r=20, t=40, b=20))
st.plotly_chart(fig_tp, use_container_width=True)

# 2️⃣ Node-wise Latency Over Time (Interactive)
st.subheader("2️⃣ Node-wise Latency Over Time")
lat = df.groupby(["second", "node_id", "action"])["duration_ms"].mean().reset_index()
lat["label"] = "Node " + lat["node_id"].astype(str) + " - " + lat["action"]
fig_lat = px.line(lat, x="second", y="duration_ms", color="label", markers=True)
fig_lat.update_layout(height=350, title="Latency Over Time (ms)", margin=dict(l=20, r=20, t=40, b=20))
st.plotly_chart(fig_lat, use_container_width=True)


st.subheader("3️⃣ Average Latency per Node")
latency_node = df.groupby(["node_id", "action"])["duration_ms"].mean().reset_index()
fig3, ax3 = compact_fig()
sns.barplot(data=latency_node, x="node_id", y="duration_ms", hue="action", ax=ax3)
ax3.set_title("Avg Latency per Node", fontsize=10)
fig3.tight_layout(pad=1)
st.pyplot(fig3, use_container_width=False)


st.subheader("4️⃣ Success Rate per Node")
success_counts = df.groupby(["node_id", "action", "success"]).size().unstack(fill_value=0)
success_counts["success_rate"] = (success_counts.get(True, 0) / success_counts.sum(axis=1)) * 100
success_counts = success_counts.reset_index()
fig4, ax4 = compact_fig()
sns.barplot(data=success_counts, x="node_id", y="success_rate", hue="action", ax=ax4)
ax4.set_title("Operation Success Rate (%)", fontsize=10)
ax4.set_ylabel("")
fig4.tight_layout(pad=1)
st.pyplot(fig4, use_container_width=False)


st.subheader("5️⃣ Total Requests per Node")
node_usage = df["node_id"].value_counts().reset_index()
node_usage.columns = ["node_id", "requests"]
fig5, ax5 = compact_fig()
sns.barplot(data=node_usage, x="node_id", y="requests", ax=ax5, palette="Set2")
ax5.set_title("Total Requests by Node", fontsize=10)
fig5.tight_layout(pad=1)
st.pyplot(fig5, use_container_width=False)


st.subheader("6️⃣ Worker-wise Operation Count")
worker_op = df.groupby(["worker_id", "action"]).size().reset_index(name="count")
fig6, ax6 = compact_fig()
sns.barplot(data=worker_op, x="worker_id", y="count", hue="action", ax=ax6)
ax6.set_title("Operations by Worker", fontsize=10)
fig6.tight_layout(pad=1)
st.pyplot(fig6, use_container_width=False)
