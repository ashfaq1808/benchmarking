import streamlit as st
import pandas as pd
import json
import matplotlib.pyplot as plt
import seaborn as sns
from datetime import datetime

st.set_page_config(layout="wide")
st.title("📊 Cassandra Benchmarking Results")

# Load data
with open("result.json") as f:
    data = json.load(f)

df = pd.DataFrame(data)
df['timestamp'] = pd.to_datetime(df['timestamp'])
df['duration_ms'] = df['duration'].str.replace('ms', '').astype(float)
df['second'] = df['timestamp'].dt.floor('S')

# Throughput
throughput_df = df.groupby(['second', 'action']).size().unstack(fill_value=0).reset_index()
st.subheader("🔁 Throughput Over Time")
st.line_chart(throughput_df.set_index('second'))

# Latency
latency_df = df.groupby(['second', 'action'])['duration_ms'].mean().unstack(fill_value=0).reset_index()
st.subheader("⏱️ Latency Over Time")
st.line_chart(latency_df.set_index('second'))

# Success vs Failure
st.subheader("✅ Success vs ❌ Failure")
success_df = df.groupby(['action', 'success']).size().unstack(fill_value=0)
st.bar_chart(success_df)
