<template>
  <div>
    <el-input placeholder="请输入内容" v-model="content" class="input-with-select"
      style="margin-top: 15px;width:50%;margin-left: 15px;">
      <el-select v-model="select" slot="prepend" placeholder="请选择">
        <el-option label="商品ID" value="1"></el-option>
        <el-option label="姓名" value="2"></el-option>
      </el-select>
      <el-button slot="append" icon="el-icon-search" @click="handleSearch()"></el-button>
    </el-input>

    <el-table :data="tableData" border style="width: 100%">
      <el-table-column label="订单ID" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.orderNum }}
        </template>
      </el-table-column>
      <el-table-column label="商品ID" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.goodId }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="价格" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.enc_b_m }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="买家" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.buyer }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="卖家" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.seller }}
          <el-popover />
        </template>
      </el-table-column>

      <el-table-column label="卖家对交易金额的承诺" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.commB }}
        </template>
      </el-table-column>
      <el-table-column label="卖家对交易金额承诺的签名" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.sign_commB }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="卖家确认交易的签名" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.sign_confirm }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="买家对交易金额的承诺" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.commA }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="买家对交易金额承诺的签名" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.sign_commA }}
          <el-popover />
        </template>
      </el-table-column>

      <el-table-column label="交易金额大于零的证据" width="" align="center">
        <template slot-scope="scope">
          <div class="scrollable-content">
            {{ scope.row.rp_m }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="交易余额大于零的证据" width="" align="center">
        <template slot-scope="scope">
          <div class="scrollable-content">
            {{ scope.row.rp_b }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="金额环签名" width="" align="center">
        <template slot-scope="scope">
          <div class="scrollable-content">
            {{ scope.row.link_sign_1 }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="身份环签名" width="" align="center">
        <template slot-scope="scope">
          <div class="scrollable-content">
            {{ scope.row.link_sign_2 }}
          </div>
        </template>
      </el-table-column>

    </el-table>
  </div>
</template>

<style>
.el-select .el-input {
  width: 130px;
}

.input-with-select .el-input-group__prepend {
  background-color: #fff;
}

.scrollable-content {
  max-height: 300px;
  /* 设置你希望的最大高度 */
  overflow-y: auto;
  /* 当内容超出时显示垂直滚动条 */
  /* 你可以添加其他样式，比如 padding, border 等 */
}
</style>
<script type="text/javascript">
import { listOrder } from '@/api/order';
export default {
  data() {
    return {
      seller: 'Bob',
      tableData: [],
      content: '',
      select: ''
    }
  },
  methods: {
    loadData() {
      var data = {
        seller: this.seller,
      }
      listOrder(data).then(
        response => {
          this.tableData = JSON.parse(response.data)
          this.$message({
            message: response.msg,
            type: 'success'
          });
        }
      );
    },
  },
  mounted: function () {
    this.loadData()
  }
}
</script>
