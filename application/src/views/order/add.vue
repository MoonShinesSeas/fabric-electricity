<template>
  <div class="">
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

      <el-table-column label="交易金额承诺" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.commB }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="交易金额承诺签名" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.sign_commB }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="确认交易的签名" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.sign_confirm }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="同态计算后的余额" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.enc_b_b }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center">
        <template slot-scope="scope">
          <!-- 使用v-if或v-show来控制按钮的显示 -->
          <el-button size="mini" type="primary" @click="handleSubmit(scope.$index, scope.row)">同意</el-button>
          <el-button size="mini" type="danger" @click="handleRefuse(scope.$index, scope.row)">拒绝</el-button>
        </template>

      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import { listOrder, submitOrder } from '@/api/order';
export default {
  data() {
    return {
      tableData: [],
      seller: "Bob",
      buyer: "Alice",
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
    handleSubmit(index,row) {
      var orderNum = row.orderNum;
      var buyer = this.buyer;
      var seller = this.seller;
      var price = row.enc_b_m
      if (orderNum === null || orderNum === ""
        || buyer === null || buyer === ""
        || seller === null || seller === ""
        || price === null || price === ""
      ) {
        this.$message({
          message: '请填写完成的信息',
          type: 'warning'
        });
      } else {
        //提交修改请求
        var data = {
          orderNum: orderNum,
          buyer: buyer,
          seller: seller,
          price: String(price),
        }
        submitOrder(data).then(
          response => {
            // this.tableData = JSON.parse(response.data)
            this.$message({
              message: response.msg,
              type: 'success'
            });
          }
        );
      }
    },
  },
  mounted: function () {
    this.loadData()
  }
}
</script>
