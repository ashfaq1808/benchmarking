import json
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import plotly.express as px
import plotly.graph_objects as go
from plotly.subplots import make_subplots
import streamlit as st
from datetime import datetime
import os

st.set_page_config(page_title="Cassandra Benchmark Dashboard", layout="wide")
st.title("📊 Cassandra Benchmark Dashboard with System Monitoring")

# Load benchmark results
@st.cache_data
def load_benchmark_data():
    try:
        with open("result.json", "r") as f:
            data = json.load(f)
        
        df = pd.json_normalize(data)
        df["timestamp"] = pd.to_datetime(df["timestamp"])
        
        def parse_duration_to_ms(duration_str):
            if pd.isna(duration_str):
                return 0.0
            
            duration_str = str(duration_str)
            
            if 'ms' in duration_str:
                return float(duration_str.replace('ms', ''))
            elif 'µs' in duration_str or 'us' in duration_str:
                return float(duration_str.replace('µs', '').replace('us', '')) / 1000.0
            elif 's' in duration_str and 'ms' not in duration_str:
                return float(duration_str.replace('s', '')) * 1000.0
            elif 'ns' in duration_str:
                return float(duration_str.replace('ns', '')) / 1000000.0
            else:
                try:
                    return float(duration_str)
                except:
                    return 0.0

        df["duration_ms"] = df["duration"].apply(parse_duration_to_ms)
        df["second"] = df["timestamp"].dt.floor("s")
        df = df[df["node_id"].notnull()] 
        
        return df
    except FileNotFoundError:
        return None

# Load system metrics
@st.cache_data
def load_system_metrics():
    try:
        with open("system_metrics.json", "r") as f:
            data = json.load(f)
        
        metrics_df = pd.json_normalize(data)
        metrics_df["timestamp"] = pd.to_datetime(metrics_df["timestamp"])
        
        return metrics_df
    except FileNotFoundError:
        return None

# Load data
benchmark_df = load_benchmark_data()
metrics_df = load_system_metrics()

# Create tabs for different views
tab1, tab2, tab3, tab4 = st.tabs(["🎯 Benchmark Results", "💻 System Performance", "🐳 Docker Metrics", "📈 Combined Analysis"])

if benchmark_df is not None:
    with tab1:
        st.header("Cassandra Benchmark Results")
        
        col1, col2, col3, col4 = st.columns(4)
        with col1:
            st.metric("Total Operations", len(benchmark_df))
        with col2:
            success_rate = (benchmark_df['success'].sum() / len(benchmark_df)) * 100
            st.metric("Success Rate", f"{success_rate:.2f}%")
        with col3:
            avg_latency = benchmark_df["duration_ms"].mean()
            st.metric("Avg Latency", f"{avg_latency:.2f}ms")
        with col4:
            total_duration = (benchmark_df["timestamp"].max() - benchmark_df["timestamp"].min()).total_seconds()
            st.metric("Test Duration", f"{total_duration:.1f}s")

        # Request Rate Pattern
        st.subheader("Request Rate Pattern")
        requests_per_second = benchmark_df.groupby(["second", "node_id"]).size().reset_index(name="requests")
        requests_per_second["node_label"] = "Node " + requests_per_second["node_id"].astype(str)

        fig_rate = px.line(requests_per_second, x="second", y="requests", color="node_label", 
                          title="Request Distribution Per Second")
        fig_rate.update_layout(height=400)
        st.plotly_chart(fig_rate, use_container_width=True)

        # Latency Analysis
        col1, col2 = st.columns(2)
        with col1:
            lat = benchmark_df.groupby(["second", "node_id", "action"])["duration_ms"].mean().reset_index()
            lat["label"] = "Node " + lat["node_id"].astype(str) + " - " + lat["action"]
            fig_lat = px.line(lat, x="second", y="duration_ms", color="label", 
                             title="Latency Over Time (ms)")
            st.plotly_chart(fig_lat, use_container_width=True)

        with col2:
            latency_node = benchmark_df.groupby(["node_id", "action"])["duration_ms"].mean().reset_index()
            fig_bar = px.bar(latency_node, x="node_id", y="duration_ms", color="action",
                            title="Average Latency per Node")
            st.plotly_chart(fig_bar, use_container_width=True)

else:
    with tab1:
        st.warning("⚠️ Benchmark results file (result.json) not found. Run a benchmark first.")

# System Performance Tab
with tab2:
    st.header("System Performance Monitoring")
    
    if metrics_df is not None:
        # Time range selector
        time_range = st.selectbox("Select Time Range", 
                                 ["Last 5 minutes", "Last 15 minutes", "Last 30 minutes", "All"])
        
        if time_range != "All":
            minutes = int(time_range.split()[1])
            cutoff_time = metrics_df["timestamp"].max() - pd.Timedelta(minutes=minutes)
            filtered_metrics = metrics_df[metrics_df["timestamp"] >= cutoff_time]
        else:
            filtered_metrics = metrics_df

        if not filtered_metrics.empty:
            # CPU Metrics
            st.subheader("🖥️ CPU Performance")
            col1, col2 = st.columns(2)
            
            with col1:
                if 'cpu.usage_percent' in filtered_metrics.columns:
                    fig_cpu = px.line(filtered_metrics, x="timestamp", y="cpu.usage_percent", 
                                     title="CPU Usage (%)")
                    fig_cpu.update_layout(height=300)
                    st.plotly_chart(fig_cpu, use_container_width=True)
                    
                    current_cpu = filtered_metrics['cpu.usage_percent'].iloc[-1] if len(filtered_metrics) > 0 else 0
                    st.metric("Current CPU Usage", f"{current_cpu:.1f}%")

            with col2:
                if 'cpu.load_average.load_1' in filtered_metrics.columns:
                    load_cols = ['cpu.load_average.load_1', 'cpu.load_average.load_5', 'cpu.load_average.load_15']
                    available_load_cols = [col for col in load_cols if col in filtered_metrics.columns]
                    
                    if available_load_cols:
                        fig_load = go.Figure()
                        for col in available_load_cols:
                            fig_load.add_trace(go.Scatter(
                                x=filtered_metrics["timestamp"], 
                                y=filtered_metrics[col],
                                name=col.split('.')[-1].replace('_', ' ').title(),
                                mode='lines+markers'
                            ))
                        fig_load.update_layout(title="Load Average", height=300)
                        st.plotly_chart(fig_load, use_container_width=True)

            # Memory Metrics
            st.subheader("🧠 Memory Performance")
            col1, col2 = st.columns(2)
            
            with col1:
                if 'memory.usage_percent' in filtered_metrics.columns:
                    fig_mem = px.line(filtered_metrics, x="timestamp", y="memory.usage_percent", 
                                     title="Memory Usage (%)")
                    fig_mem.update_layout(height=300)
                    st.plotly_chart(fig_mem, use_container_width=True)
                    
                    current_mem = filtered_metrics['memory.usage_percent'].iloc[-1] if len(filtered_metrics) > 0 else 0
                    st.metric("Current Memory Usage", f"{current_mem:.1f}%")

            with col2:
                memory_cols = ['memory.used_bytes', 'memory.free_bytes', 'memory.cached_bytes']
                available_mem_cols = [col for col in memory_cols if col in filtered_metrics.columns]
                
                if available_mem_cols:
                    fig_mem_detail = go.Figure()
                    for col in available_mem_cols:
                        values = filtered_metrics[col] / (1024**3)  # Convert to GB
                        fig_mem_detail.add_trace(go.Scatter(
                            x=filtered_metrics["timestamp"], 
                            y=values,
                            name=col.split('.')[-1].replace('_', ' ').title(),
                            mode='lines'
                        ))
                    fig_mem_detail.update_layout(title="Memory Breakdown (GB)", height=300)
                    st.plotly_chart(fig_mem_detail, use_container_width=True)

            # Network Metrics
            st.subheader("🌐 Network Performance")
            col1, col2 = st.columns(2)
            
            with col1:
                if 'network.total_rx_rate_mbps' in filtered_metrics.columns:
                    fig_net_rx = px.line(filtered_metrics, x="timestamp", y="network.total_rx_rate_mbps", 
                                        title="Network RX Rate (Mbps)")
                    fig_net_rx.update_layout(height=300)
                    st.plotly_chart(fig_net_rx, use_container_width=True)

            with col2:
                if 'network.total_tx_rate_mbps' in filtered_metrics.columns:
                    fig_net_tx = px.line(filtered_metrics, x="timestamp", y="network.total_tx_rate_mbps", 
                                        title="Network TX Rate (Mbps)")
                    fig_net_tx.update_layout(height=300)
                    st.plotly_chart(fig_net_tx, use_container_width=True)

            # GPU Metrics (if available)
            gpu_cols = [col for col in filtered_metrics.columns if 'gpu.devices' in col]
            if gpu_cols:
                st.subheader("🎮 GPU Performance")
                
                # Extract GPU metrics
                gpu_usage_cols = [col for col in filtered_metrics.columns if 'gpu.devices.0.usage_percent' in col]
                gpu_memory_cols = [col for col in filtered_metrics.columns if 'gpu.devices.0.memory_usage_percent' in col]
                gpu_temp_cols = [col for col in filtered_metrics.columns if 'gpu.devices.0.temperature_celsius' in col]
                
                col1, col2, col3 = st.columns(3)
                
                with col1:
                    if gpu_usage_cols:
                        fig_gpu_usage = px.line(filtered_metrics, x="timestamp", y=gpu_usage_cols[0], 
                                               title="GPU Usage (%)")
                        fig_gpu_usage.update_layout(height=250)
                        st.plotly_chart(fig_gpu_usage, use_container_width=True)

                with col2:
                    if gpu_memory_cols:
                        fig_gpu_mem = px.line(filtered_metrics, x="timestamp", y=gpu_memory_cols[0], 
                                             title="GPU Memory Usage (%)")
                        fig_gpu_mem.update_layout(height=250)
                        st.plotly_chart(fig_gpu_mem, use_container_width=True)

                with col3:
                    if gpu_temp_cols:
                        fig_gpu_temp = px.line(filtered_metrics, x="timestamp", y=gpu_temp_cols[0], 
                                              title="GPU Temperature (°C)")
                        fig_gpu_temp.update_layout(height=250)
                        st.plotly_chart(fig_gpu_temp, use_container_width=True)

    else:
        st.warning("⚠️ System metrics file (system_metrics.json) not found. Enable monitoring in benchmark configuration.")

# Docker Metrics Tab
with tab3:
    st.header("Docker Container Performance")
    
    if metrics_df is not None:
        docker_cols = [col for col in metrics_df.columns if 'docker.' in col]
        
        if docker_cols:
            # Time range selector
            time_range = st.selectbox("Select Time Range", 
                                     ["Last 5 minutes", "Last 15 minutes", "Last 30 minutes", "All"],
                                     key="docker_time_range")
            
            if time_range != "All":
                minutes = int(time_range.split()[1])
                cutoff_time = metrics_df["timestamp"].max() - pd.Timedelta(minutes=minutes)
                filtered_metrics = metrics_df[metrics_df["timestamp"] >= cutoff_time]
            else:
                filtered_metrics = metrics_df

            col1, col2 = st.columns(2)
            
            with col1:
                st.subheader("Container CPU & Memory")
                if 'docker.cpu.usage_percent' in filtered_metrics.columns:
                    fig_docker_cpu = px.line(filtered_metrics, x="timestamp", y="docker.cpu.usage_percent", 
                                            title="Container CPU Usage (%)")
                    fig_docker_cpu.update_layout(height=300)
                    st.plotly_chart(fig_docker_cpu, use_container_width=True)

                if 'docker.memory.percent' in filtered_metrics.columns:
                    fig_docker_mem = px.line(filtered_metrics, x="timestamp", y="docker.memory.percent", 
                                            title="Container Memory Usage (%)")
                    fig_docker_mem.update_layout(height=300)
                    st.plotly_chart(fig_docker_mem, use_container_width=True)

            with col2:
                st.subheader("Container Network & I/O")
                if 'docker.network.rx_bytes' in filtered_metrics.columns and 'docker.network.tx_bytes' in filtered_metrics.columns:
                    fig_docker_net = go.Figure()
                    fig_docker_net.add_trace(go.Scatter(
                        x=filtered_metrics["timestamp"], 
                        y=filtered_metrics["docker.network.rx_bytes"] / (1024**2),
                        name="RX (MB)", mode='lines'
                    ))
                    fig_docker_net.add_trace(go.Scatter(
                        x=filtered_metrics["timestamp"], 
                        y=filtered_metrics["docker.network.tx_bytes"] / (1024**2),
                        name="TX (MB)", mode='lines'
                    ))
                    fig_docker_net.update_layout(title="Container Network I/O (MB)", height=300)
                    st.plotly_chart(fig_docker_net, use_container_width=True)

                if 'docker.block_io.read_bytes' in filtered_metrics.columns and 'docker.block_io.write_bytes' in filtered_metrics.columns:
                    fig_docker_io = go.Figure()
                    fig_docker_io.add_trace(go.Scatter(
                        x=filtered_metrics["timestamp"], 
                        y=filtered_metrics["docker.block_io.read_bytes"] / (1024**2),
                        name="Read (MB)", mode='lines'
                    ))
                    fig_docker_io.add_trace(go.Scatter(
                        x=filtered_metrics["timestamp"], 
                        y=filtered_metrics["docker.block_io.write_bytes"] / (1024**2),
                        name="Write (MB)", mode='lines'
                    ))
                    fig_docker_io.update_layout(title="Container Block I/O (MB)", height=300)
                    st.plotly_chart(fig_docker_io, use_container_width=True)

            # Container info
            if 'docker.container_name' in filtered_metrics.columns:
                container_name = filtered_metrics['docker.container_name'].iloc[0] if len(filtered_metrics) > 0 else "Unknown"
                container_id = filtered_metrics['docker.container_id'].iloc[0] if 'docker.container_id' in filtered_metrics.columns and len(filtered_metrics) > 0 else "Unknown"
                st.info(f"📦 Container: {container_name} (ID: {container_id[:12]})")

        else:
            st.warning("⚠️ No Docker metrics found. Enable Docker monitoring in benchmark configuration.")
    else:
        st.warning("⚠️ System metrics file not found. Enable monitoring in benchmark configuration.")

# Combined Analysis Tab
with tab4:
    st.header("Combined Performance Analysis")
    
    if benchmark_df is not None and metrics_df is not None:
        st.subheader("Correlation Analysis")
        
        # Merge benchmark and metrics data by timestamp
        benchmark_df['timestamp_floor'] = benchmark_df['timestamp'].dt.floor('5S')  # 5-second intervals
        metrics_df['timestamp_floor'] = metrics_df['timestamp'].dt.floor('5S')
        
        # Aggregate benchmark data
        benchmark_agg = benchmark_df.groupby('timestamp_floor').agg({
            'duration_ms': 'mean',
            'success': 'mean'
        }).reset_index()
        
        # Aggregate system metrics
        system_cols = ['cpu.usage_percent', 'memory.usage_percent']
        available_system_cols = [col for col in system_cols if col in metrics_df.columns]
        
        if available_system_cols:
            metrics_agg = metrics_df.groupby('timestamp_floor')[available_system_cols].mean().reset_index()
            
            # Merge datasets
            combined_df = pd.merge(benchmark_agg, metrics_agg, on='timestamp_floor', how='inner')
            
            if not combined_df.empty:
                # Create correlation matrix
                corr_cols = ['duration_ms'] + available_system_cols
                correlation_matrix = combined_df[corr_cols].corr()
                
                col1, col2 = st.columns(2)
                
                with col1:
                    fig_corr = px.imshow(correlation_matrix, 
                                        title="Performance Correlation Matrix",
                                        aspect="auto")
                    st.plotly_chart(fig_corr, use_container_width=True)
                
                with col2:
                    # Time series overlay
                    fig_overlay = make_subplots(
                        rows=2, cols=1,
                        subplot_titles=('Latency vs CPU Usage', 'Latency vs Memory Usage'),
                        vertical_spacing=0.1
                    )
                    
                    if 'cpu.usage_percent' in combined_df.columns:
                        fig_overlay.add_trace(
                            go.Scatter(x=combined_df['timestamp_floor'], y=combined_df['duration_ms'], 
                                     name='Latency (ms)', yaxis='y'),
                            row=1, col=1
                        )
                        fig_overlay.add_trace(
                            go.Scatter(x=combined_df['timestamp_floor'], y=combined_df['cpu.usage_percent'], 
                                     name='CPU Usage (%)', yaxis='y2'),
                            row=1, col=1
                        )
                    
                    if 'memory.usage_percent' in combined_df.columns:
                        fig_overlay.add_trace(
                            go.Scatter(x=combined_df['timestamp_floor'], y=combined_df['duration_ms'], 
                                     name='Latency (ms)', showlegend=False),
                            row=2, col=1
                        )
                        fig_overlay.add_trace(
                            go.Scatter(x=combined_df['timestamp_floor'], y=combined_df['memory.usage_percent'], 
                                     name='Memory Usage (%)', yaxis='y4'),
                            row=2, col=1
                        )
                    
                    fig_overlay.update_layout(height=500, title="Performance Correlation Over Time")
                    st.plotly_chart(fig_overlay, use_container_width=True)
                
                # Performance insights
                st.subheader("📊 Performance Insights")
                
                col1, col2, col3 = st.columns(3)
                
                with col1:
                    avg_latency = combined_df['duration_ms'].mean()
                    max_latency = combined_df['duration_ms'].max()
                    st.metric("Average Latency", f"{avg_latency:.2f}ms")
                    st.metric("Peak Latency", f"{max_latency:.2f}ms")
                
                with col2:
                    if 'cpu.usage_percent' in combined_df.columns:
                        avg_cpu = combined_df['cpu.usage_percent'].mean()
                        max_cpu = combined_df['cpu.usage_percent'].max()
                        st.metric("Average CPU", f"{avg_cpu:.1f}%")
                        st.metric("Peak CPU", f"{max_cpu:.1f}%")
                
                with col3:
                    if 'memory.usage_percent' in combined_df.columns:
                        avg_memory = combined_df['memory.usage_percent'].mean()
                        max_memory = combined_df['memory.usage_percent'].max()
                        st.metric("Average Memory", f"{avg_memory:.1f}%")
                        st.metric("Peak Memory", f"{max_memory:.1f}%")
                
                # Recommendations
                st.subheader("💡 Performance Recommendations")
                
                if 'cpu.usage_percent' in combined_df.columns:
                    high_cpu_threshold = 80
                    high_cpu_periods = combined_df[combined_df['cpu.usage_percent'] > high_cpu_threshold]
                    if len(high_cpu_periods) > 0:
                        st.warning(f"⚠️ CPU usage exceeded {high_cpu_threshold}% for {len(high_cpu_periods)} time periods. Consider scaling up CPU resources.")
                
                if 'memory.usage_percent' in combined_df.columns:
                    high_mem_threshold = 85
                    high_mem_periods = combined_df[combined_df['memory.usage_percent'] > high_mem_threshold]
                    if len(high_mem_periods) > 0:
                        st.warning(f"⚠️ Memory usage exceeded {high_mem_threshold}% for {len(high_mem_periods)} time periods. Consider increasing memory allocation.")
                
                latency_threshold = 100  # ms
                high_latency_periods = combined_df[combined_df['duration_ms'] > latency_threshold]
                if len(high_latency_periods) > 0:
                    st.warning(f"⚠️ Latency exceeded {latency_threshold}ms for {len(high_latency_periods)} time periods. This may indicate performance bottlenecks.")
                
            else:
                st.info("No overlapping data found between benchmark results and system metrics.")
    else:
        st.warning("⚠️ Both benchmark results and system metrics are required for combined analysis.")

# Footer
st.markdown("---")
st.markdown("📊 **Cassandra Benchmark Dashboard** - Real-time performance monitoring and analysis")
if metrics_df is not None:
    st.markdown(f"📈 System metrics collected: {len(metrics_df)} data points")
if benchmark_df is not None:
    st.markdown(f"🎯 Benchmark operations: {len(benchmark_df)} operations recorded")